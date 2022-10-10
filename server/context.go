// SPDX-License-Identifier: MIT

package server

import (
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

	xencoding "github.com/issue9/web/internal/encoding"
	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/serializer"
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

// Context 根据当次 HTTP 请求生成的上下文内容
//
// Context 同时也实现了 [http.ResponseWriter] 接口，
// 但是不推荐非必要情况下直接使用 [http.ResponseWriter] 的接口方法，
// 而是采用返回 [Responser] 的方式向客户端输出内容。
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

func (ctx *Context) Route() types.Route { return ctx.route }

// SetLanguage 修改输出的语言
func (ctx *Context) SetLanguage(tag language.Tag) {
	if ctx.languageTag != tag {
		ctx.languageTag = tag
		ctx.localePrinter = ctx.Server().NewPrinter(tag)
	}
}

func (ctx *Context) LocalePrinter() *message.Printer { return ctx.localePrinter }

func (ctx *Context) LanguageTag() language.Tag { return ctx.languageTag }

// SetLocation 设置时区信息
func (ctx *Context) SetLocation(loc *time.Location) { ctx.location = loc }

func (ctx *Context) Location() *time.Location { return ctx.location }

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

// Logs 返回日志管理对象
func (ctx *Context) Logs() *logs.Logs { return ctx.Server().Logs() }

func (ctx *Context) IsXHR() bool {
	h := strings.ToLower(ctx.Request().Header.Get("X-Requested-With"))
	return h == "xmlhttprequest"
}
