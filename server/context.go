// SPDX-License-Identifier: MIT

package server

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v4"
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

const (
	contextPoolBodyBufferMaxSize = 1 << 16
	defaultBodyBufferSize        = 256
)

var contextPool = &sync.Pool{New: func() any { return &Context{} }}

// Context 根据当次 HTTP 请求生成的上下文内容
//
// Context 同时也实现了 http.ResponseWriter 接口，
// 但是不推荐非必要情况下直接使用 http.ResponseWriter 的接口方法，
// 而是采用返回 Responser 的方式向客户端输出内容。
type Context struct {
	server      *Server
	params      params.Params
	request     *http.Request
	contentType string
	exits       []func(int)

	// response
	resp           http.ResponseWriter // 原始的 http.ResponseWriter
	respWriter     io.Writer           // http.ResponseWriter.Write 实际写入的对象
	encodingCloser io.WriteCloser
	charsetCloser  io.WriteCloser
	outputEncoding *serialization.EncodingBuilder
	outputCharset  encoding.Encoding
	status         int // http.ResponseWriter.WriteHeader 保存的副本
	wrote          bool

	// 指定将 Response 输出时所使用的媒体类型。从 Accept 报头解析得到。
	// 如果是调用 Context.Write 输出内容，可以为空。
	outputMimetype serialization.MarshalFunc

	// 从客户端提交的 Content-Type 报头解析到的内容
	inputMimetype serialization.UnmarshalFunc // 可以为空
	inputCharset  encoding.Encoding           // 若值为 encoding.Nop 或是 nil，表示为 utf-8

	// 区域和本地相关信息
	languageTag   language.Tag
	localePrinter *message.Printer
	location      *time.Location

	body []byte // 缓存从 http.Request.Body 中获取的内容
	read bool   // 表示是已经读取 body

	// 保存 Context 在存续期间的可复用变量
	//
	// 这是比 context.Value 更经济的传递变量方式，但是这并不是协程安全的。
	Vars map[any]any
}

// 如果出错，则会向 w 输出状态码并返回 nil。
func (srv *Server) newContext(w http.ResponseWriter, r *http.Request, ps params.Params) *Context {
	header := r.Header.Get("Accept")
	outputMimetypeName, marshal, found := srv.mimetypes.MarshalFunc(header)
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
	outputEncoding, notAcceptable := srv.encodings.Search(outputMimetypeName, header)
	if notAcceptable {
		w.WriteHeader(http.StatusNotAcceptable)
		return nil
	}

	tag := srv.acceptLanguage(r.Header.Get("Accept-Language"))

	header = r.Header.Get("Content-Type")
	inputMimetype, inputCharset, err := srv.conentType(header)
	if err != nil {
		srv.Logs().Debug(err)
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return nil
	}

	// NOTE: ctx 是从对象池中获取的，所有变量都必须初始化。

	ctx := contextPool.Get().(*Context)
	ctx.server = srv
	ctx.params = ps
	ctx.request = r
	ctx.contentType = buildContentType(outputMimetypeName, outputCharsetName)
	if len(ctx.exits) > 0 {
		ctx.exits = ctx.exits[:0]
	}

	// response
	ctx.resp = w
	ctx.respWriter = w
	ctx.encodingCloser = nil
	ctx.charsetCloser = nil
	ctx.outputEncoding = outputEncoding
	ctx.outputCharset = outputCharset
	ctx.status = http.StatusOK // 需是 http.StatusOK，否则在未调用 WriteHeader 的情况下会与默认情况不符。
	ctx.wrote = false
	if ctx.outputEncoding != nil {
		h := ctx.Header()
		h.Set("Content-Encoding", ctx.outputEncoding.Name())
		h.Add("Vary", "Content-Encoding")
	}

	ctx.outputMimetype = marshal
	ctx.inputMimetype = inputMimetype
	ctx.inputCharset = inputCharset
	ctx.languageTag = tag
	ctx.localePrinter = srv.Locale().NewPrinter(tag)
	ctx.location = srv.location
	if len(ctx.body) > 0 {
		ctx.body = ctx.body[:0]
	}
	ctx.read = false
	ctx.Vars = make(map[any]any)

	return ctx
}

func (ctx *Context) Write(bs []byte) (int, error) {
	if !ctx.wrote { // 在第一次有内容输出时，才决定构建 Encoding 和 Charset 的 io.Writer
		ctx.wrote = true

		if ctx.outputEncoding != nil {
			ctx.encodingCloser = ctx.outputEncoding.Build(ctx.respWriter)
			ctx.respWriter = ctx.encodingCloser
		}

		if !charsetIsNop(ctx.outputCharset) {
			ctx.charsetCloser = transform.NewWriter(ctx.respWriter, ctx.outputCharset.NewEncoder())
			ctx.respWriter = ctx.charsetCloser
		}
	}

	return ctx.respWriter.Write(bs)
}

func (ctx *Context) WriteHeader(status int) {
	ctx.Header().Del("Content-Length") // https://github.com/golang/go/issues/14975
	ctx.status = status
	ctx.resp.WriteHeader(status)
}

func (ctx *Context) Header() http.Header { return ctx.resp.Header() }

// SetLanguage 设置输出的语言
//
// 默认情况下，会根据用户提交的 Accept-Language 报头设置默认值。
func (ctx *Context) SetLanguage(l string) error {
	tag, err := language.Parse(l)
	if err == nil {
		ctx.languageTag = tag
		ctx.localePrinter = ctx.Server().Locale().NewPrinter(tag)
	}
	return err
}

func (ctx *Context) LocalePrinter() *message.Printer { return ctx.localePrinter }

func (ctx *Context) LanguageTag() language.Tag { return ctx.languageTag }

// SetLocation 设置时区信息
//
// name 为时区名称，比如 'America/New_York'，具体说明可参考 time.LoadLocataion
func (ctx *Context) SetLocation(name string) error {
	l, err := time.LoadLocation(name)
	if err == nil {
		ctx.location = l
	}
	return err
}

func (ctx *Context) Location() *time.Location { return ctx.location }

// Request 返回原始的请求对象
//
// NOTE: 如果需要使用 context.Value 在 Request 中附加值，请使用 ctx.Vars
func (ctx *Context) Request() *http.Request { return ctx.request }

func (ctx *Context) destroy() {
	if ctx.charsetCloser != nil {
		if err := ctx.charsetCloser.Close(); err != nil {
			ctx.Logs().ERROR().Error(err) // 出错记录日志但不退出，之后的 Exit 还是要调用
		}
	}

	if ctx.encodingCloser != nil { // encoding 在最底层，应该最后关闭。
		if err := ctx.encodingCloser.Close(); err != nil {
			ctx.Logs().ERROR().Error(err) // 出错记录日志但不退出，之后的 Exit 还是要调用
		}
	}

	for _, exit := range ctx.exits {
		exit(ctx.status)
	}

	if len(ctx.body) < contextPoolBodyBufferMaxSize { // 过大的对象不回收，以免造成内存占用过高。
		contextPool.Put(ctx)
	}
}

// OnExit 注册退出当前请求时的处理函数
//
// f 的原型为
//  func(status int)
// 其中 status 为最终输出到客户端的状态码。
func (ctx *Context) OnExit(f func(int)) {
	if ctx.exits == nil {
		ctx.exits = make([]func(int), 0, 3)
	}
	ctx.exits = append(ctx.exits, f)
}

// Body 获取用户提交的内容
//
// 相对于 ctx.Request().Body，此函数可多次读取。不存在 body 时，返回 nil
func (ctx *Context) Body() (body []byte, err error) {
	if ctx.read {
		return ctx.body, nil
	}

	var reader io.Reader = ctx.Request().Body
	if !charsetIsNop(ctx.inputCharset) {
		reader = transform.NewReader(reader, ctx.inputCharset.NewDecoder())
	}

	if ctx.body == nil {
		ctx.body = make([]byte, 0, defaultBodyBufferSize)
	}

	for {
		if len(ctx.body) == cap(ctx.body) {
			ctx.body = append(ctx.body, 0)[:len(ctx.body)]
		}
		n, err := reader.Read(ctx.body[len(ctx.body):cap(ctx.body)])
		ctx.body = ctx.body[:len(ctx.body)+n]
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
	}

	ctx.read = true
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

// Marshal 向客户端输出内容
//
// status 想输出给用户状态码，如果出错，那么最终展示给用户的状态码可能不是此值；
// body 表示输出的对象，该对象最终调用 ctx.outputMimetype 编码；
func (ctx *Context) Marshal(status int, body any) error {
	if body == nil {
		ctx.WriteHeader(status)
		return nil
	}

	// 如果 outputMimetype 为空，说明在 Server.Mimetypes() 的配置中就是 nil。
	// 那么不应该执行到此，比如下载文件等直接从 ResponseWriter.Write 输出的。
	if ctx.outputMimetype == nil {
		ctx.WriteHeader(http.StatusNotAcceptable)
		return localeutil.Error("%s can not be empty", "ctx.outputMimetype")
	}

	ctx.Header().Set("Content-Type", ctx.contentType)
	if id := ctx.languageTag.String(); id != "" {
		ctx.Header().Set("Content-Language", id)
	}

	data, err := ctx.outputMimetype(body)
	switch {
	case errors.Is(err, serialization.ErrUnsupported):
		ctx.WriteHeader(http.StatusNotAcceptable)
		return err
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

	f, found := srv.mimetypes.UnmarshalFunc(mt)
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
func (ctx *Context) Now() time.Time { return time.Now().In(ctx.Location()) }

// ParseTime 分析基于当前时区的时间
func (ctx *Context) ParseTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, ctx.Location())
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
	return ctx.LocalePrinter().Sprintf(key, v...)
}
