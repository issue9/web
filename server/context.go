// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v3"
	"github.com/issue9/mux/v6/params"
	"github.com/issue9/qheader"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/transform"

	"github.com/issue9/web/serialization"
)

// 在 sync.Pool 回收 Context 时，如果 body 长度超过此值，则不回收，以免造成占用过高的内存。
const poolContextBodyMaxSize = 1 << 16

var (
	// 需要作比较，所以得是经过 http.CanonicalHeaderKey 处理的标准名称。
	contentTypeKey     = http.CanonicalHeaderKey("Content-Type")
	contentLanguageKey = http.CanonicalHeaderKey("Content-Language")

	contextPool = &sync.Pool{New: func() any { return &Context{} }}
)

// CTXSanitizer 提供对数据的验证和修正
//
// 在 Context.Read 和 Queries.Object 中会在解析数据成功之后，调用该接口进行数据验证。
type CTXSanitizer interface {
	// CTXSanitize 验证和修正当前对象的数据
	//
	// 如果验证有误，则需要返回这些错误信息。
	CTXSanitize(*Context) ResultFields
}

// Context 根据当次 HTTP 请求生成的上下文内容
//
// Context 同时也实现了 http.ResponseWriter 接口，
// 但是不推荐非必要情况下直接使用 http.ResponseWriter 的接口方法，
// 而是采用返回 Response 的方式向客户端输出内容。
type Context struct {
	server            *Server
	params            params.Params
	outputCharsetName string
	request           *http.Request

	// http.ResponseWriter
	encodingCloser io.WriteCloser
	charsetCloser  io.WriteCloser
	resp           http.ResponseWriter
	respWriter     io.Writer
	rendered       bool

	// 指定将 Response 输出时所使用的媒体类型，以及名称。从 Accept 报头解析得到。
	outputMimetype     serialization.MarshalFunc // 如果是调用 Context.Write 输出内容，可以为空。
	outputMimetypeName string                    // 如果为空，则采用 DefaultMimetype

	// 从客户端提交的 Content-Type 报头解析到的内容
	inputMimetype serialization.UnmarshalFunc // 可以为空
	inputCharset  encoding.Encoding           // 若值为 encoding.Nop 或是 nil，表示为 utf-8

	// 输出语言的相关设置项
	OutputTag     language.Tag
	LocalePrinter *message.Printer

	Location *time.Location

	body []byte // 缓存从 http.Request.Body 中获取的内容
	read bool   // 表示是已经读取 body

	// 保存 Context 在存续期间的可复用变量
	//
	// 这是比 context.Value 更经济的传递变量方式，但是这并不是协程安全的。
	Vars map[any]any
}

// NewContext 构建 *Context 实例
//
// 如果不合规则，则会向 w 输出状态码并返回 nil。
func (srv *Server) NewContext(w http.ResponseWriter, r *http.Request) *Context {
	header := r.Header.Get("Accept")
	outputMimetypeName, marshal, found := srv.Mimetypes().MarshalFunc(header)
	if !found {
		srv.Logs().Debug(srv.localePrinter.Sprintf("not found serialization for %s", header))
		w.WriteHeader(http.StatusNotAcceptable)
		return nil
	}

	header = r.Header.Get("Accept-Charset")
	outputCharsetName, outputCharset := acceptCharset(header)
	if outputCharsetName == "" {
		srv.Logs().Debug(srv.localePrinter.Sprintf("not found charset for %s", header))
		w.WriteHeader(http.StatusNotAcceptable)
		return nil
	}

	header = r.Header.Get("Accept-Encoding")
	outputEncoding, notAcceptable := srv.Encodings().Search(outputMimetypeName, header)
	if notAcceptable {
		w.WriteHeader(http.StatusNotAcceptable)
		return nil
	}

	tag := srv.acceptLanguage(r.Header.Get("Accept-Language"))

	header = r.Header.Get(contentTypeKey)
	inputMimetype, inputCharset, err := srv.conentType(header)
	if err != nil {
		srv.Logs().Debug(err)
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return nil
	}

	// NOTE: ctx 是从对象池中获取的，必须所有变量都初始化
	ctx := contextPool.Get().(*Context)
	ctx.server = srv
	ctx.params = nil
	ctx.outputCharsetName = outputCharsetName
	ctx.request = r

	// 初始化 encodingCloser, charsetCloser, resp, respWriter, rendered
	srv.buildResponse(w, ctx, outputCharset, outputEncoding)

	ctx.outputMimetype = marshal
	ctx.outputMimetypeName = outputMimetypeName
	ctx.inputMimetype = inputMimetype
	ctx.inputCharset = inputCharset
	ctx.OutputTag = tag
	ctx.LocalePrinter = srv.Locale().NewPrinter(tag)
	ctx.Location = srv.location
	ctx.body = ctx.body[:0]
	ctx.read = false
	ctx.Vars = make(map[any]any)
	return ctx
}

func (ctx *Context) Write(bs []byte) (int, error) {
	ctx.rendered = true
	return ctx.respWriter.Write(bs)
}

func (ctx *Context) WriteHeader(status int) {
	ctx.rendered = true
	ctx.resp.WriteHeader(status)
}

func (ctx *Context) Header() http.Header { return ctx.resp.Header() }

// Request 返回原始的请求对象
//
// NOTE: 如果需要使用 context.Value 在 Request 中附加值，请使用 ctx.Vars
func (ctx *Context) Request() *http.Request { return ctx.request }

func (srv *Server) buildResponse(resp http.ResponseWriter, ctx *Context, c encoding.Encoding, b *serialization.EncodingBuilder) {
	ctx.resp = resp
	ctx.respWriter = resp
	ctx.encodingCloser = nil
	ctx.charsetCloser = nil
	ctx.rendered = false

	h := resp.Header()

	if b != nil {
		ctx.encodingCloser = b.Build(ctx.respWriter)
		ctx.respWriter = ctx.encodingCloser
		h.Del("Content-Length") // https://github.com/golang/go/issues/14975
		h.Set("Content-Encoding", b.Name())
		h.Add("Vary", "Content-Encoding")
	}

	if !charsetIsNop(c) {
		ctx.charsetCloser = transform.NewWriter(ctx.respWriter, c.NewEncoder())
		ctx.respWriter = ctx.charsetCloser
	}
}

func (ctx *Context) destroy() error {
	if ctx.charsetCloser != nil {
		if err := ctx.charsetCloser.Close(); err != nil {
			return err
		}
	}

	if ctx.encodingCloser != nil { // encoding 在最底层，应该最后关闭。
		if err := ctx.encodingCloser.Close(); err != nil {
			return err
		}
	}

	// 过大的对象不回收，以免造成内存占用过高。
	if len(ctx.body) < poolContextBodyMaxSize {
		contextPool.Put(ctx)
	}

	return nil
}

// Body 获取用户提交的内容
//
// 相对于 ctx.Request().Body，此函数可多次读取。不存在 body 时，返回 nil
func (ctx *Context) Body() (body []byte, err error) {
	if ctx.read {
		return ctx.body, nil
	}

	if ctx.body, err = io.ReadAll(ctx.Request().Body); err != nil {
		return nil, err
	}
	ctx.read = true

	if charsetIsNop(ctx.inputCharset) {
		return ctx.body, nil
	}

	d := ctx.inputCharset.NewDecoder()
	reader := transform.NewReader(bytes.NewReader(ctx.body), d)
	ctx.body, err = io.ReadAll(reader)
	return ctx.body, err
}

// Unmarshal 将提交的内容转换成 v 对象
func (ctx *Context) Unmarshal(v any) error {
	body, err := ctx.Body()
	if err != nil {
		return err
	}

	if len(body) == 0 {
		return nil
	}
	return ctx.inputMimetype(body, v)
}

func (ctx *Context) Marshal(status int, body any, headers map[string]string) error {
	if ctx.rendered {
		return localeutil.Error("rendered")
	}

	header := ctx.Header()

	var contentTypeFound, contentLanguageFound bool
	for k, v := range headers {
		k = http.CanonicalHeaderKey(k)

		contentTypeFound = contentTypeFound || k == contentTypeKey
		contentLanguageFound = contentLanguageFound || k == contentLanguageKey
		header.Set(k, v)
	}

	if body == nil {
		ctx.WriteHeader(status)
		return nil
	}

	// 如果 outputMimetype 为空，那么不应该执行到此，比如下载文件等直接从 ResponseWriter.Write 输出的，需要另行处理。
	if ctx.outputMimetype == nil {
		ctx.WriteHeader(http.StatusNotAcceptable)
		return nil
	}

	if !contentTypeFound {
		ct := buildContentType(ctx.outputMimetypeName, ctx.outputCharsetName)
		header.Set(contentTypeKey, ct)
	}

	if !contentLanguageFound && ctx.OutputTag != language.Und {
		header.Set(contentLanguageKey, ctx.OutputTag.String())
	}

	data, err := ctx.outputMimetype(body)
	switch {
	case errors.Is(err, serialization.ErrUnsupported):
		ctx.WriteHeader(http.StatusNotAcceptable)
		return nil
	case err != nil:
		ctx.WriteHeader(http.StatusInternalServerError)
		return err
	}

	ctx.WriteHeader(status)
	_, err = ctx.Write(data)
	return err
}

// 根据 Accept-Charset 报头的内容获取其最值的字符集信息
//
// 传递 * 获取返回默认的字符集相关信息，即 utf-8
// 其它值则按值查找，或是在找不到时返回空值。
//
// 返回的 name 值可能会与 header 中指定的不一样，比如 gb_2312 会被转换成 gbk
func acceptCharset(header string) (name string, enc encoding.Encoding) {
	if header == "" || header == "*" {
		return DefaultCharset, nil
	}

	var err error
	accepts := qheader.Parse(header, "*")
	for _, apt := range accepts {
		enc, err = htmlindex.Get(apt.Value)
		if err != nil { // err != nil 表示未找到，继续查找
			continue
		}

		// 转换成官方的名称
		name, err = htmlindex.Name(enc)
		if err != nil {
			name = apt.Value // 不存在，直接使用用户上传的名称
		}

		return name, enc
	}

	return "", nil
}

func (srv *Server) acceptLanguage(header string) language.Tag {
	if header == "" {
		return srv.Tag()
	}
	tag, _ := language.MatchStrings(srv.locale.Builder().Matcher(), header)
	return tag
}

func (srv *Server) conentType(header string) (serialization.UnmarshalFunc, encoding.Encoding, error) {
	var mt, charset = DefaultMimetype, DefaultCharset

	if header != "" {
		m, ps, err := mime.ParseMediaType(header)
		if err != nil {
			return nil, nil, err
		}
		mt = m

		if c := ps["charset"]; c != "" {
			charset = c
		}
	}

	f, found := srv.Mimetypes().UnmarshalFunc(mt)
	if !found {
		return nil, nil, localeutil.Error("not found serialization function for %s", mt)
	}

	e, err := htmlindex.Get(charset)
	if err != nil {
		return nil, nil, err
	}

	return f, e, nil
}

// Now 返回以 ctx.Location 为时区的当前时间
func (ctx *Context) Now() time.Time { return time.Now().In(ctx.Location) }

// ParseTime 分析基于当前时区的时间
func (ctx *Context) ParseTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, ctx.Location)
}

// Read 从客户端读取数据并转换成 v 对象
//
// 功能与 Unmarshal() 相同，只不过 Read() 在出错时，返回的不是 error，
// 而是一个表示错误信息的 Response 对象。
//
// 如果 v 实现了 CTXSanitizer 接口，则在读取数据之后，会调用其接口函数。
// 如果验证失败，会输出以 code 作为错误代码的 Response 对象。
func (ctx *Context) Read(v any, code string) Responser {
	if err := ctx.Unmarshal(v); err != nil {
		return ctx.Error(http.StatusUnprocessableEntity, err)
	}

	if vv, ok := v.(CTXSanitizer); ok {
		if rslt := vv.CTXSanitize(ctx); len(rslt) > 0 {
			return ctx.Result(code, rslt)
		}
	}

	return nil
}

// ClientIP 返回客户端的 IP 地址及端口
//
// 获取顺序如下：
//  - X-Forwarded-For 的第一个元素
//  - Remote-Addr 报头
//  - X-Read-IP 报头
func (ctx *Context) ClientIP() string {
	ip := ctx.Request().Header.Get("X-Forwarded-For")
	if index := strings.IndexByte(ip, ','); index > 0 {
		ip = ip[:index]
	}
	if ip == "" && ctx.Request().RemoteAddr != "" {
		ip = ctx.Request().RemoteAddr
	}
	if ip == "" {
		ip = ctx.Request().Header.Get("X-Real-IP")
	}

	return strings.TrimSpace(ip)
}

func (ctx *Context) Logs() *logs.Logs { return ctx.Server().Logs() }

// Log 输出日志并以指定的状态码退出
//
// deep 为 0 表示 Log 本身；
func (ctx *Context) Log(level, deep int, v ...any) {
	ctx.Logs().Print(level, deep, v...)
}

// Logf 输出日志并以指定的状态码退出
//
// deep 为 0 表示 Logf 本身；
func (ctx *Context) Logf(level, deep int, format string, v ...any) {
	ctx.Logs().Printf(level, deep, format, v...)
}

// 指定的编码是否不需要任何额外操作
func charsetIsNop(enc encoding.Encoding) bool {
	return enc == nil || enc == unicode.UTF8 || enc == encoding.Nop
}

// 生成 content-type，若值为空，则会使用默认值代替。
func buildContentType(mt, charset string) string {
	if mt == "" {
		mt = DefaultMimetype
	}
	if charset == "" {
		charset = DefaultCharset
	}

	return mt + "; charset=" + charset
}

func (ctx *Context) IsXHR() bool {
	h := strings.ToLower(ctx.Request().Header.Get("X-Requested-With"))
	return h == "xmlhttprequest"
}

// Sprintf 返回翻译后的结果
func (ctx *Context) Sprintf(key message.Reference, v ...any) string {
	return ctx.LocalePrinter.Sprintf(key, v...)
}
