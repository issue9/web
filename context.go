// SPDX-License-Identifier: MIT

package web

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/issue9/mux/v7/types"
	"golang.org/x/text/encoding"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/web/internal/compress"
	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/logs"
)

const contextPoolBodyBufferMaxSize = 1 << 16 // 过大的对象不回收，以免造成内存占用过高。

var contextPool = &sync.Pool{
	New: func() any {
		return &Context{exits: make([]func(*Context, int), 0, 5)}
	},
}

// Context 根据当次 HTTP 请求生成的上下文内容
//
// Context 同时也实现了 [http.ResponseWriter] 接口，
// 但是不推荐非必要情况下直接使用 [http.ResponseWriter] 的接口方法，
// 而是采用返回 [Responser] 的方式向客户端输出内容。
type Context struct {
	server            *Server
	route             types.Route
	request           *http.Request
	outputCharsetName string
	exits             []func(*Context, int)
	id                string
	begin             time.Time
	keepAlive         bool

	// response
	originResponse http.ResponseWriter // 原始的 http.ResponseWriter
	writer         io.Writer           // 实际写入的对象
	encodingCloser io.WriteCloser
	charsetCloser  io.WriteCloser
	outputEncoding *compress.NamedCompress
	outputCharset  encoding.Encoding
	status         int // http.ResponseWriter.WriteHeader 保存的副本
	wrote          bool

	// 指定将 Response 输出时所使用的媒体类型。从 Accept 报头解析得到。
	// 如果是调用 Context.Write 输出内容，outputMimetype.Marshal 可以为空。
	outputMimetype *mtType

	// 从客户端提交的 Content-Type 报头解析到的内容
	inputMimetype UnmarshalFunc     // 可以为空
	inputCharset  encoding.Encoding // 若值为 encoding.Nop 或是 nil，表示为 utf-8

	// 区域和本地相关信息
	languageTag   language.Tag
	localePrinter *message.Printer

	requestBody []byte // 缓存从 http.Request.Body 中获取的内容
	read        bool   // 表示是已经读取 body

	// 保存 Context 在存续期间的可复用变量
	//
	// 这是比 [context.Value] 更经济的传递变量方式，但是这并不是协程安全的。
	vars map[any]any

	logs Logs
}

// 如果出错，则会向 w 输出状态码并返回 nil。
func (srv *Server) newContext(w http.ResponseWriter, r *http.Request, route types.Route) *Context {
	id := buildID(srv, w, r)
	l := srv.Logs().With(map[string]any{
		srv.requestIDKey: id,
	})

	h := r.Header.Get(header.Accept)
	mt := srv.mimetypes.Accept(h)
	if mt == nil {
		l.DEBUG().String(Phrase("not found serialization for %s", h).LocaleString(srv.LocalePrinter()))
		w.WriteHeader(http.StatusNotAcceptable)
		return nil
	}

	h = r.Header.Get(header.AcceptCharset)
	outputCharsetName, outputCharset := header.ParseAcceptCharset(h)
	if outputCharsetName == "" {
		l.DEBUG().String(Phrase("not found charset for %s", h).LocaleString(srv.LocalePrinter()))
		w.WriteHeader(http.StatusNotAcceptable)
		return nil
	}

	h = r.Header.Get(header.AcceptEncoding)
	outputEncoding, notAcceptable := srv.compresses.AcceptEncoding(mt.Name, h, l.DEBUG())
	if notAcceptable {
		w.WriteHeader(http.StatusNotAcceptable)
		return nil
	}

	tag := srv.acceptLanguage(r.Header.Get(header.AcceptLang))

	var inputMimetype UnmarshalFunc
	var inputCharset encoding.Encoding
	h = r.Header.Get(header.ContentType)
	if h != "" {
		var err error
		inputMimetype, inputCharset, err = srv.mimetypes.ContentType(h)
		if err != nil {
			l.DEBUG().Printf("%+v", err)
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return nil
		}
	}

	// NOTE: ctx 是从对象池中获取的，所有变量都必须初始化。

	ctx := contextPool.Get().(*Context)
	ctx.server = srv
	ctx.route = route
	ctx.request = r
	ctx.outputCharsetName = outputCharsetName
	ctx.exits = ctx.exits[:0]
	ctx.id = id
	ctx.begin = time.Now()
	ctx.keepAlive = false

	// response
	ctx.originResponse = w
	ctx.writer = w
	ctx.encodingCloser = nil
	ctx.charsetCloser = nil
	ctx.outputEncoding = outputEncoding
	ctx.outputCharset = outputCharset
	ctx.status = 0
	ctx.wrote = false
	if ctx.outputEncoding != nil {
		h := ctx.Header()
		h.Set(header.ContentEncoding, ctx.outputEncoding.Name())
		h.Add(header.Vary, header.ContentEncoding)
	}

	ctx.outputMimetype = mt
	ctx.inputMimetype = inputMimetype
	ctx.inputCharset = inputCharset
	ctx.languageTag = tag
	ctx.localePrinter = srv.NewLocalePrinter(tag)
	if len(ctx.requestBody) > 0 {
		ctx.requestBody = ctx.requestBody[:0]
	}
	ctx.read = false
	ctx.vars = map[any]any{} // TODO: go1.21 可以改为 clear(ctx.vars)

	ctx.logs = l

	return ctx
}

// NewContext 从标准库的参数初始化 Context 对象
//
// NOTE: 这适合从标准库的请求中创建 [Context] 对象，
// 但是部分功能会缺失，比如地址中的参数信息，以及 [Context.Route] 等。
func (srv *Server) NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return srv.newContext(w, r, types.NewContext())
}

// GetVar 返回指定名称的变量
func (ctx *Context) GetVar(key any) (any, bool) {
	v, found := ctx.vars[key]
	return v, found
}

// SetVar 设置变量
func (ctx *Context) SetVar(key, val any) { ctx.vars[key] = val }

// Route 关联的路由信息
func (ctx *Context) Route() types.Route { return ctx.route }

func buildID(s *Server, w http.ResponseWriter, r *http.Request) string {
	id := r.Header.Get(s.requestIDKey)
	if id == "" {
		id = s.UniqueID()
	}

	w.Header().Set(s.requestIDKey, id)
	return id
}

// Begin 当前对象的初始化时间
func (ctx *Context) Begin() time.Time { return ctx.begin }

// KeepAlive 保持 Context 在路由处理函数外依然可用
//
// 默认情况下，Context 在当前路由结束时会被放回到对象池中以备下次使用，
// 此方法可以跳过该操作，但依然会被 Go 的 GC 正常回收。
func (ctx *Context) KeepAlive() { ctx.keepAlive = true }

// ID 当前请求的唯一 ID
//
// 一般源自客户端的 X-Request-ID 报头，如果不存在，则由 [Server.UniqueID] 生成。
func (ctx *Context) ID() string { return ctx.id }

// SetCharset 设置输出的字符集
//
// 相当于重新设置了 [Context.Request] 的 Accept-Charset 报头，但是不会实际修改 [Context.Request]。
func (ctx *Context) SetCharset(charset string) {
	if ctx.Wrote() {
		panic("已有内容输出，不可再更改！")
	}
	if ctx.Charset() == charset {
		return
	}

	outputCharsetName, outputCharset := header.ParseAcceptCharset(charset)
	if outputCharsetName == "" {
		panic(fmt.Sprintf("指定的字符集 %s 不存在", charset))
	}
	ctx.outputCharset = outputCharset
	ctx.outputCharsetName = outputCharsetName
}

// Charset 输出的字符集
func (ctx *Context) Charset() string { return ctx.outputCharsetName }

// SetMimetype 设置输出的格式
//
// 相当于重新设置了 [Context.Request] 的 Accept 报头，但是不会实际修改 [Context.Request]。
func (ctx *Context) SetMimetype(mimetype string) {
	if ctx.Wrote() {
		panic("已有内容输出，不可再更改！")
	}
	if ctx.Mimetype(false) == mimetype {
		return
	}

	item := ctx.Server().mimetypes.Accept(mimetype)
	if item == nil {
		panic(fmt.Sprintf("指定的编码 %s 不存在", mimetype))
	}
	ctx.outputMimetype = item
}

// Mimetype 返回输出编码名称
//
// problem 表示是否返回 problem 状态时的值。该值由 [Mimetype.ProblemType] 设置。
func (ctx *Context) Mimetype(problem bool) string {
	if ctx.outputMimetype == nil {
		return ""
	}

	if problem {
		return ctx.outputMimetype.Problem
	}
	return ctx.outputMimetype.Name
}

// SetEncoding 设置输出的压缩编码
func (ctx *Context) SetEncoding(enc string) {
	if ctx.Wrote() {
		panic("已有内容输出，不可再更改！")
	}
	if ctx.Encoding() == enc {
		return
	}

	outputEncoding, notAcceptable := ctx.Server().compresses.AcceptEncoding(ctx.outputMimetype.Name, enc, ctx.Logs().DEBUG())
	if notAcceptable {
		panic(fmt.Sprintf("指定的压缩编码 %s 不存在", enc))
	}
	ctx.outputEncoding = outputEncoding
}

// Encoding 输出的压缩编码名称
func (ctx *Context) Encoding() string {
	if ctx.outputEncoding == nil {
		return ""
	}
	return ctx.outputEncoding.Name()
}

// SetLanguage 修改输出的语言
//
// 相当于重新设置了 [Context.Request] 的 Accept-Language 报头，但是不会实际修改 [Context.Request]。
func (ctx *Context) SetLanguage(tag language.Tag) {
	// 不判断是否有内容已经输出，允许中途改变语言。
	if ctx.languageTag != tag {
		ctx.languageTag = tag
		ctx.localePrinter = ctx.Server().NewLocalePrinter(tag)
	}
}

func (ctx *Context) LocalePrinter() *message.Printer { return ctx.localePrinter }

func (ctx *Context) LanguageTag() language.Tag { return ctx.languageTag }

func (ctx *Context) destroy() {
	if ctx.charsetCloser != nil {
		if err := ctx.charsetCloser.Close(); err != nil {
			ctx.Logs().ERROR().Printf("%+v", err) // 出错记录日志但不退出，之后的 Exit 还是要调用
		}
	}

	if ctx.encodingCloser != nil { // encoding 在最底层，应该最后关闭。
		if err := ctx.encodingCloser.Close(); err != nil {
			ctx.Logs().ERROR().Printf("%+v", err) // 出错记录日志但不退出，之后的 Exit 还是要调用
		}
	}

	for _, exit := range ctx.exits {
		exit(ctx, ctx.status)
	}

	logs.DestroyWithLogs(ctx.logs)

	if !ctx.keepAlive && len(ctx.requestBody) < contextPoolBodyBufferMaxSize {
		contextPool.Put(ctx)
	}
}

// OnExit 注册退出当前请求时的处理函数
func (ctx *Context) OnExit(f func(*Context, int)) { ctx.exits = append(ctx.exits, f) }

func (srv *Server) acceptLanguage(header string) language.Tag {
	if header == "" {
		return srv.Language()
	}
	tag, _ := language.MatchStrings(srv.Catalog().Matcher(), header)
	return tag
}

// ClientIP 返回客户端的 IP 地址及端口
//
// 获取顺序如下：
//   - X-Forwarded-For 的第一个元素
//   - Remote-Addr 报头
//   - X-Read-IP 报头
func (ctx *Context) ClientIP() string { return header.ClientIP(ctx.Request()) }

// Logs 返回日志操作对象
//
// 该返回实例与 [Server.Logs] 是不同的，
// 当前返回实例的日志输出时会带上当前请求的 [Context.ID] 作为额外参数。
func (ctx *Context) Logs() Logs { return ctx.logs }

func (ctx *Context) IsXHR() bool {
	return strings.ToLower(ctx.Request().Header.Get("X-Requested-With")) == "xmlhttprequest"
}

// Unwrap [http.ResponseController] 通过此方法返回底层的 [http.ResponseWriter]
func (ctx *Context) Unwrap() http.ResponseWriter { return ctx.originResponse }
