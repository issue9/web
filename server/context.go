// SPDX-License-Identifier: MIT

package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/issue9/logs/v4"
	"github.com/issue9/mux/v7/types"
	"golang.org/x/text/encoding"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/transform"

	xencoding "github.com/issue9/web/internal/encoding"
	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/serializer"
	"github.com/issue9/web/validation"
)

const (
	contextPoolBodyBufferMaxSize = 1 << 16
	defaultBodyBufferSize        = 256
)

var contextPool = &sync.Pool{New: func() any {
	return &Context{
		exits: make([]func(int), 0, 3), // query, params
	}
}}

// CTXSanitizer 在 [Context] 关联的上下文环境中提供对数据的验证和修正
//
// 在 [Context.Read] 和 [Queries.Object] 中会在解析数据成功之后，调用该接口进行数据验证。
type CTXSanitizer interface {
	// CTXSanitize 验证和修正当前对象的数据
	//
	// 如果验证有误，则需要返回这些错误信息，否则应该返回 nil。
	CTXSanitize(*Context) *validation.Validation
}

// Context 根据当次 HTTP 请求生成的上下文内容
//
// Context 同时也实现了 [http.ResponseWriter] 接口，
// 但是不推荐非必要情况下直接使用 [http.ResponseWriter] 的接口方法，
// 而是采用返回 Responser 的方式向客户端输出内容。
type Context struct {
	server             *Server
	route              types.Route
	request            *http.Request
	outputMimetypeName string
	outputCharsetName  string
	exits              []func(int)

	// response
	resp           http.ResponseWriter // 原始的 http.ResponseWriter
	respWriter     io.Writer           // http.ResponseWriter.Write 实际写入的对象
	encodingCloser io.WriteCloser
	charsetCloser  io.WriteCloser
	outputEncoding *xencoding.Pool
	outputCharset  encoding.Encoding
	status         int // http.ResponseWriter.WriteHeader 保存的副本
	wrote          bool

	// 指定将 Response 输出时所使用的媒体类型。从 Accept 报头解析得到。
	// 如果是调用 Context.Write 输出内容，可以为空。
	outputMimetype serializer.MarshalFunc

	// 从客户端提交的 Content-Type 报头解析到的内容
	inputMimetype serializer.UnmarshalFunc // 可以为空
	inputCharset  encoding.Encoding        // 若值为 encoding.Nop 或是 nil，表示为 utf-8

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
func (srv *Server) newContext(w http.ResponseWriter, r *http.Request, route types.Route) *Context {
	h := r.Header.Get("Accept")
	outputMimetypeName, marshal, found := srv.mimetypes.MarshalFunc(h)
	if !found {
		srv.Logs().Debug(srv.LocalePrinter().Sprintf("not found serialization for %s", h))
		w.WriteHeader(http.StatusNotAcceptable)
		return nil
	}

	h = r.Header.Get("Accept-Charset")
	outputCharsetName, outputCharset := header.AcceptCharset(h)
	if outputCharsetName == "" {
		srv.Logs().Debug(srv.LocalePrinter().Sprintf("not found charset for %s", h))
		w.WriteHeader(http.StatusNotAcceptable)
		return nil
	}

	h = r.Header.Get("Accept-Encoding")
	outputEncoding, notAcceptable := srv.encodings.Search(outputMimetypeName, h)
	if notAcceptable {
		w.WriteHeader(http.StatusNotAcceptable)
		return nil
	}

	tag := srv.acceptLanguage(r.Header.Get("Accept-Language"))

	var inputMimetype serializer.UnmarshalFunc
	var inputCharset encoding.Encoding
	h = r.Header.Get("Content-Type")
	if h != "" {
		var err error
		inputMimetype, inputCharset, err = srv.mimetypes.ContentType(h)
		if err != nil {
			srv.Logs().Debug(err)
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return nil
		}
	}

	// NOTE: ctx 是从对象池中获取的，所有变量都必须初始化。

	ctx := contextPool.Get().(*Context)
	ctx.server = srv
	ctx.route = route
	ctx.request = r
	ctx.outputMimetypeName = outputMimetypeName
	ctx.outputCharsetName = outputCharsetName
	ctx.exits = ctx.exits[:0]

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
	ctx.localePrinter = srv.NewPrinter(tag)
	ctx.location = srv.Location()
	if len(ctx.body) > 0 {
		ctx.body = ctx.body[:0]
	}
	ctx.read = false
	ctx.Vars = map[any]any{}

	return ctx
}

func (ctx *Context) Write(bs []byte) (int, error) {
	if !ctx.wrote { // 在第一次有内容输出时，才决定构建 Encoding 和 Charset 的 io.Writer
		ctx.wrote = true

		if ctx.outputEncoding != nil {
			ctx.encodingCloser = ctx.outputEncoding.Get(ctx.respWriter)
			ctx.respWriter = ctx.encodingCloser
		}

		if !header.CharsetIsNop(ctx.outputCharset) {
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

func (ctx *Context) Route() types.Route { return ctx.route }

// SetLanguage 设置输出的语言
//
// 默认情况下，会根据用户提交的 Accept-Language 报头设置默认值。
func (ctx *Context) SetLanguage(l string) error {
	tag, err := language.Parse(l)
	if err == nil {
		ctx.languageTag = tag
		ctx.localePrinter = ctx.Server().NewPrinter(tag)
	}
	return err
}

func (ctx *Context) LocalePrinter() *message.Printer { return ctx.localePrinter }

func (ctx *Context) LanguageTag() language.Tag { return ctx.languageTag }

// SetLocation 设置时区信息
//
// name 为时区名称，比如 'America/New_York'，具体说明可参考 [time.LoadLocation]
func (ctx *Context) SetLocation(name string) error {
	l, err := time.LoadLocation(name)
	if err == nil {
		ctx.location = l
	}
	return err
}

func (ctx *Context) Location() *time.Location { return ctx.location }

// Request 返回原始的请求对象
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
// f 为退出时的处理方法，其原型为：
//
//	func(status int)
//
// 其中 status 为最终输出到客户端的状态码。
func (ctx *Context) OnExit(f func(int)) {
	ctx.exits = append(ctx.exits, f)
}

// Marshal 向客户端输出内容
//
// status 想输出给用户状态码，如果出错，那么最终展示给用户的状态码可能不是此值；
// body 表示输出的对象，该对象最终调用 ctx.outputMimetype 编码；
// problem 表示 body 是否为 [Problem] 对象，对于 Problem 对象可能会有特殊的处理；
func (ctx *Context) Marshal(status int, body any, problem bool) error {
	if body == nil {
		ctx.WriteHeader(status)
		return nil
	}

	// 如果 outputMimetype 为空，说明在 Server.Mimetypes() 的配置中就是 nil。
	// 那么不应该执行到此，比如下载文件等直接从 ResponseWriter.Write 输出的。
	if ctx.outputMimetype == nil {
		ctx.WriteHeader(http.StatusNotAcceptable)
		panic(fmt.Sprintf("未对 %s 作处理", ctx.outputMimetypeName))
	}

	if problem {
		ctx.outputMimetypeName = ctx.Server().Problems().mimetype(ctx.outputMimetypeName)
	}
	ctx.Header().Set("Content-Type", header.BuildContentType(ctx.outputMimetypeName, ctx.outputCharsetName))
	if id := ctx.languageTag.String(); id != "" {
		ctx.Header().Set("Content-Language", id)
	}

	data, err := ctx.outputMimetype(body)
	switch {
	case errors.Is(err, serializer.ErrUnsupported):
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

func (srv *Server) acceptLanguage(header string) language.Tag {
	if header == "" {
		return srv.LanguageTag()
	}
	tag, _ := language.MatchStrings(srv.CatalogBuilder().Matcher(), header)
	return tag
}

// Now 返回以 [Context.Location] 为时区的当前时间
func (ctx *Context) Now() time.Time { return time.Now().In(ctx.Location()) }

// ParseTime 分析基于当前时区的时间
func (ctx *Context) ParseTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, ctx.Location())
}

// ClientIP 返回客户端的 IP 地址及端口
//
// 获取顺序如下：
//   - X-Forwarded-For 的第一个元素
//   - Remote-Addr 报头
//   - X-Read-IP 报头
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

func (ctx *Context) IsXHR() bool {
	h := strings.ToLower(ctx.Request().Header.Get("X-Requested-With"))
	return h == "xmlhttprequest"
}

// Sprintf 返回翻译后的结果
func (ctx *Context) Sprintf(key message.Reference, v ...any) string {
	return ctx.LocalePrinter().Sprintf(key, v...)
}
