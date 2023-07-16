// SPDX-License-Identifier: MIT

package server

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

	xencoding "github.com/issue9/web/internal/encoding"
	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/internal/mimetypes"
	"github.com/issue9/web/logs"
)

const contextPoolBodyBufferMaxSize = 1 << 16

var contextPool = &sync.Pool{
	New: func() any {
		return &Context{exits: make([]func(*Context, int), 0, 3)} // query, params, filter
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

	// response
	originResponse http.ResponseWriter // 原始的 http.ResponseWriter
	writer         io.Writer           // 实际写入的对象
	encodingCloser io.WriteCloser
	charsetCloser  io.WriteCloser
	outputEncoding *xencoding.Alg
	outputCharset  encoding.Encoding
	status         int // http.ResponseWriter.WriteHeader 保存的副本
	wrote          bool

	// 指定将 Response 输出时所使用的媒体类型。从 Accept 报头解析得到。
	// 如果是调用 Context.Write 输出内容，可以为空。
	outputMimetype *mimetypes.Mimetype[MarshalFunc, UnmarshalFunc]

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
	// 这是比 context.Value 更经济的传递变量方式，但是这并不是协程安全的。
	vars map[any]any

	logs *logs.ParamsLogs
}

// MarshalFunc 序列化函数原型
//
// NOTE: MarshalFunc 的实现中不能调用 [Context.Render] 方法。
type MarshalFunc func(*Context, any) ([]byte, error)

// UnmarshalFunc 反序列化函数原型
type UnmarshalFunc func([]byte, any) error

// 如果出错，则会向 w 输出状态码并返回 nil。
func (srv *Server) newContext(w http.ResponseWriter, r *http.Request, route types.Route) *Context {
	h := r.Header.Get(header.Accept)
	mt := srv.mimetypes.Accept(h)
	if mt == nil {
		srv.Logs().DEBUG().Printf("not found serialization for %s", h)
		w.WriteHeader(http.StatusNotAcceptable)
		return nil
	}

	h = r.Header.Get(header.AcceptCharset)
	outputCharsetName, outputCharset := header.ParseAcceptCharset(h)
	if outputCharsetName == "" {
		srv.Logs().DEBUG().Printf("not found charset for %s", h)
		w.WriteHeader(http.StatusNotAcceptable)
		return nil
	}

	h = r.Header.Get(header.AcceptEncoding)
	outputEncoding, notAcceptable := srv.encodings.Search(mt.Name, h)
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
			srv.Logs().DEBUG().Error(err)
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
	ctx.id = buildID(ctx.Server(), w, r)
	ctx.begin = time.Now()

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
	ctx.localePrinter = srv.NewPrinter(tag)
	if len(ctx.requestBody) > 0 {
		ctx.requestBody = ctx.requestBody[:0]
	}
	ctx.read = false
	ctx.vars = map[any]any{}

	// 在最后，保证已经存在 ctx.id 变量
	ctx.logs = srv.Logs().With(map[string]any{
		srv.requestIDKey: ctx.ID(),
	})

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

// Mimetype 输出编码名称
//
// problem 表示是否返回 problem 时的 mimetype 值。该值由 [Mimetypes] 设置。
func (ctx *Context) Mimetype(problem bool) string {
	if ctx.outputMimetype == nil {
		return ""
	}

	if problem {
		return ctx.outputMimetype.Problem
	}
	return ctx.outputMimetype.Name
}

// SetEncoding 设置压缩编码
//
// 相当于重新设置了 [Context.Request] 的 Accept-Encoding 报头，但是不会实际修改 [Context.Request]。
func (ctx *Context) SetEncoding(enc string) {
	if ctx.Wrote() {
		panic("已有内容输出，不可再更改！")
	}
	if ctx.Encoding() == enc {
		return
	}

	outputEncoding, notAcceptable := ctx.Server().encodings.Search(ctx.outputMimetype.Name, enc)
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
		ctx.localePrinter = ctx.Server().NewPrinter(tag)
	}
}

func (ctx *Context) LocalePrinter() *message.Printer { return ctx.localePrinter }

func (ctx *Context) LanguageTag() language.Tag { return ctx.languageTag }

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
		exit(ctx, ctx.status)
	}

	logs.DestroyParamsLogs(ctx.logs)

	if len(ctx.requestBody) < contextPoolBodyBufferMaxSize { // 过大的对象不回收，以免造成内存占用过高。
		contextPool.Put(ctx)
	}
}

// OnExit 注册退出当前请求时的处理函数
func (ctx *Context) OnExit(f func(*Context, int)) { ctx.exits = append(ctx.exits, f) }

func (srv *Server) acceptLanguage(header string) language.Tag {
	if header == "" {
		return srv.Language()
	}
	tag, _ := language.MatchStrings(srv.CatalogBuilder().Matcher(), header)
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
// 当前返回实例的日志输出时会带上当前请求的 x-request-id 作为额外参数。
//
// 输出内容依然遵照 [Server.Logs] 的规则作本地化处理。
func (ctx *Context) Logs() *logs.ParamsLogs { return ctx.logs }

func (ctx *Context) IsXHR() bool {
	h := strings.ToLower(ctx.Request().Header.Get("X-Requested-With"))
	return h == "xmlhttprequest"
}

// Unwrap 返回底层的 http.ResponseWriter
//
// 在 go1.20 之后，可由 [http.ResponseController] 可能要用到此方法。
func (ctx *Context) Unwrap() http.ResponseWriter { return ctx.originResponse }
