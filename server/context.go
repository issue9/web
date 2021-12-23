// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v3"
	"github.com/issue9/qheader"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/transform"

	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/serialization"
)

// 需要作比较，所以得是经过 http.CanonicalHeaderKey 处理的标准名称。
var (
	contentTypeKey     = http.CanonicalHeaderKey("Content-Type")
	contentLanguageKey = http.CanonicalHeaderKey("Content-Language")
)

var contextPool = &sync.Pool{
	New: func() interface{} { return &Context{} },
}

// CTXSanitizer 提供对数据的验证和修正
//
// 但凡对象实现了该接口，那么在 Context.Read 和 Queries.Object
// 中会在解析数据成功之后，调用该接口进行数据验证。
//
// 可用于 HTTP 请求中对用户提交数据的验证。
type CTXSanitizer interface {
	// CTXSanitize 验证和修正当前对象的数据
	//
	// 返回的是字段名以及对应的错误信息，一个字段可以对应多个错误信息。
	CTXSanitize(*Context) ResultFields
}

// Context 是对当次 HTTP 请求内容的封装
type Context struct {
	server *Server

	Response http.ResponseWriter
	Request  *http.Request

	// 指定输出时所使用的媒体类型，以及名称
	OutputMimetype     serialization.MarshalFunc
	OutputMimetypeName string

	// 输出到客户端的字符集
	//
	// 若值为 encoding.Nop 或是空，表示为 utf-8
	OutputCharset     encoding.Encoding
	OutputCharsetName string

	// 客户端内容所使用的媒体类型
	InputMimetype serialization.UnmarshalFunc

	// 客户端内容所使用的字符集
	//
	// 若值为 encoding.Nop 或是空，表示为 utf-8
	InputCharset encoding.Encoding

	// 输出语言的相关设置项
	OutputTag     language.Tag
	LocalePrinter *message.Printer

	// 与当前对话相关的时区
	Location *time.Location

	// 保存着从 http.Request.Body 中获取的内容。
	//
	// body 用于缓存从 http.Request.Body 中读取的内容；
	// read 表示是否需要从 http.Request.Body 读取内容。
	body []byte
	read bool

	// 保存 Context 在存续期间的可复用变量
	//
	// 这是比 context.Value 更经济的传递变量方式。
	//
	// 如果需要在多个请求中传递参数，可直接使用 Server.Vars。
	Vars map[interface{}]interface{}
}

// NewContext 构建 *Context 实例
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	if ctx := r.Context().Value(contextKeyContext); ctx != nil {
		return ctx.(*Context)
	}
	return GetServer(r).NewContext(w, r)
}

// NewContext 构建 *Context 实例
//
// 如果不合规则，会以指定的状码退出。
// 比如 Accept 的内容与当前配置无法匹配，则退出(panic)并输出 NotAcceptable 状态码。
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

	header = r.Header.Get(contentTypeKey)
	inputMimetype, inputCharset, err := srv.conentType(header)
	if err != nil {
		srv.Logs().Debug(err)
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return nil
	}

	tag := srv.acceptLanguage(r.Header.Get("Accept-Language"))

	// NOTE: ctx 是从对象池中获取的，必须所有变量都初始化
	ctx := contextPool.Get().(*Context)
	ctx.server = srv
	ctx.Response = w
	ctx.Request = r
	ctx.OutputMimetype = marshal
	ctx.OutputMimetypeName = outputMimetypeName
	ctx.OutputCharset = outputCharset
	ctx.OutputCharsetName = outputCharsetName
	ctx.InputMimetype = inputMimetype
	ctx.InputCharset = inputCharset
	ctx.OutputTag = tag
	ctx.LocalePrinter = srv.Locale().Printer(tag)
	ctx.Location = srv.location
	ctx.body = ctx.body[:0]
	ctx.read = false
	ctx.Vars = make(map[interface{}]interface{})

	ctx.Request = r.WithContext(context.WithValue(r.Context(), contextKeyContext, ctx))
	return ctx
}

// Body 获取用户提交的内容
//
// 相对于 ctx.Request.Body，此函数可多次读取。不存在 body 时，返回 nil
func (ctx *Context) Body() (body []byte, err error) {
	if ctx.read {
		return ctx.body, nil
	}

	if ctx.body, err = io.ReadAll(ctx.Request.Body); err != nil {
		return nil, err
	}
	ctx.read = true

	if charsetIsNop(ctx.InputCharset) {
		return ctx.body, nil
	}

	d := ctx.InputCharset.NewDecoder()
	reader := transform.NewReader(bytes.NewReader(ctx.body), d)
	ctx.body, err = io.ReadAll(reader)
	return ctx.body, err
}

// Unmarshal 将提交的内容转换成 v 对象
func (ctx *Context) Unmarshal(v interface{}) error {
	body, err := ctx.Body()
	if err != nil {
		return err
	}

	if ctx.InputMimetype != nil {
		return ctx.InputMimetype(body, v)
	}
	return nil
}

// Marshal 将 v 解码并发送给客户端
//
// status 表示输出的状态码，如果出错，可能输出的不是该状态码；
// v 输出的对象，若是一个 nil 值，则不会向客户端输出任何内容；
// 若是需要正常输出一个 nil 类型到客户端（比如JSON 中的 null），
// 可以传递一个 *struct{} 值，或是自定义实现相应的解码函数；
// headers 报头信息，如果已经存在于 ctx.Response 将覆盖 ctx.Response 中的值，
// 如果需要指定一个特定的 Content-Type 和 Content-Language，
// 可以在 headers 中指定，否则使用当前的编码和语言名称；
func (ctx *Context) Marshal(status int, v interface{}, headers map[string]string) error {
	header := ctx.Response.Header()
	var contentTypeFound, contentLanguageFound bool
	for k, v := range headers {
		k = http.CanonicalHeaderKey(k)

		contentTypeFound = contentTypeFound || k == contentTypeKey
		contentLanguageFound = contentLanguageFound || k == contentLanguageKey
		header.Set(k, v)
	}

	if !contentTypeFound {
		ct := buildContentType(ctx.OutputMimetypeName, ctx.OutputCharsetName)
		header.Set(contentTypeKey, ct)
	}

	if !contentLanguageFound && ctx.OutputTag != language.Und {
		header.Set(contentLanguageKey, ctx.OutputTag.String())
	}

	if v == nil {
		ctx.Response.WriteHeader(status)
		return nil
	}

	// 没有指定编码函数，NewContext 阶段是允许 OutputMimetype 为空的，所以只能在此处判断。
	if ctx.OutputMimetype == nil {
		ctx.Response.WriteHeader(http.StatusNotAcceptable)
		return nil
	}
	data, err := ctx.OutputMimetype(v)
	switch {
	case errors.Is(err, serialization.ErrUnsupported):
		ctx.Response.WriteHeader(http.StatusNotAcceptable)
		return nil
	case err != nil:
		ctx.Response.WriteHeader(http.StatusInternalServerError)
		return err
	}

	// 注意 WriteHeader 调用顺序。
	// https://github.com/golang/go/issues/17083
	ctx.Response.WriteHeader(status)

	if charsetIsNop(ctx.OutputCharset) {
		_, err = ctx.Response.Write(data)
		return err
	}

	w := transform.NewWriter(ctx.Response, ctx.OutputCharset.NewEncoder())
	defer func() {
		err = errs.Merge(err, w.Close())
	}()

	_, err = w.Write(data)
	return err
}

// acceptCharset 根据 Accept-Charset 报头的内容获取其最值的字符集信息
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

	al := qheader.Parse(header, "*")
	tags := make([]language.Tag, 0, len(al))
	for _, l := range al {
		tags = append(tags, language.Make(l.Value))
	}

	tag, _, _ := srv.locale.Builder().Matcher().Match(tags...)
	return tag
}

// 从 content-type 报头解析出需要用到的解码函数
func (srv *Server) conentType(header string) (serialization.UnmarshalFunc, encoding.Encoding, error) {
	var (
		mt      = DefaultMimetype
		charset = DefaultCharset
	)

	if header != "" {
		m, params, err := mime.ParseMediaType(header)
		if err != nil {
			return nil, nil, err
		}
		mt = m

		if c := params["charset"]; c != "" {
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

// Now 返回当前时间
//
// 与 time.Now() 的区别在于 Now() 基于当前时区
func (ctx *Context) Now() time.Time { return time.Now().In(ctx.Location) }

// ParseTime 分析基于当前时区的时间
func (ctx *Context) ParseTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, ctx.Location)
}

// Read 从客户端读取数据并转换成 v 对象
//
// 功能与 Unmarshal() 相同，只不过 Read() 在出错时，返回的不是 error，
// 而是一个表示错误信息的 Responser 对象。
//
// 如果 v 实现了 CTXSanitizer 接口，则在读取数据之后，会调用其接口函数。
// 如果验证失败，会输出以 code 作为错误代码的 Responser 对象。
func (ctx *Context) Read(v interface{}, code string) (err Responser) {
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

// ClientIP 返回客户端的 IP 地址
//
// NOTE: 包含了端口部分。
//
// 获取顺序如下：
//  - X-Forwarded-For 的第一个元素
//  - Remote-Addr 报头
//  - X-Read-IP 报头
func (ctx *Context) ClientIP() string {
	ip := ctx.Request.Header.Get("X-Forwarded-For")
	if index := strings.IndexByte(ip, ','); index > 0 {
		ip = ip[:index]
	}
	if ip == "" && ctx.Request.RemoteAddr != "" {
		ip = ctx.Request.RemoteAddr
	}
	if ip == "" {
		ip = ctx.Request.Header.Get("X-Real-IP")
	}

	return strings.TrimSpace(ip)
}

func (ctx *Context) Logs() *logs.Logs { return ctx.Server().Logs() }

// Log 输出日志并以指定的状态码退出
//
// deep 为 0 表示 Log 本身；
func (ctx *Context) Log(level, deep, status int, v ...interface{}) {
	ctx.Logs().Print(level, deep, v...)
}

// Logf 输出日志并以指定的状态码退出
//
// deep 为 0 表示 Logf 本身；
func (ctx *Context) Logf(level, deep, status int, format string, v ...interface{}) {
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
	h := strings.ToLower(ctx.Request.Header.Get("X-Requested-With"))
	return h == "xmlhttprequest"
}

// Sprintf 返回翻译后的结果
func (ctx *Context) Sprintf(key message.Reference, v ...interface{}) string {
	return ctx.LocalePrinter.Sprintf(key, v...)
}
