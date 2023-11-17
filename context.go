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
	"github.com/issue9/sliceutil"
	"golang.org/x/text/encoding"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/logs"
)

var contextPool = &sync.Pool{
	New: func() any { return &Context{exits: make([]func(*Context, int), 0, 5)} },
}

// Context 根据当次 HTTP 请求生成的上下文内容
//
// Context 同时也实现了 [http.ResponseWriter] 接口，
// 但是不推荐非必要情况下直接使用 [http.ResponseWriter] 的接口方法，
// 而是采用返回 [Responser] 的方式向客户端输出内容。
type Context struct {
	b       *ContextBuilder
	route   types.Route
	request *http.Request
	exits   []func(*Context, int)
	id      string
	begin   time.Time
	queries *Queries

	originResponse    http.ResponseWriter // 原始的 http.ResponseWriter
	writer            io.Writer
	outputCompressor  Compressor
	outputCharset     encoding.Encoding
	outputCharsetName string
	status            int // WriteHeader 保存的副本
	wrote             bool

	// 输出时所使用的编码类型。一般从 Accept 报头解析得到。
	// 如果是调用 Context.Write 输出内容，outputMimetype.Marshal 可以为空。
	outputMimetype *mimetype

	// 从客户端提交的 Content-Type 报头解析到的内容
	inputMimetype UnmarshalFunc // 可以为空
	inputCharset  encoding.Encoding

	// 区域和本地相关信息
	languageTag   language.Tag
	localePrinter *message.Printer

	// 保存 Context 在存续期间的可复用变量
	//
	// 这是比 [context.Value] 更经济的传递变量方式，但是这并不是协程安全的。
	vars map[any]any

	logs Logs
}

// ContextBuilder 用于创建 [Context]
type ContextBuilder struct {
	server       Server
	requestIDKey string
	codec        *Codec
}

// NewContextBuilder 声明 [ContextBuilder]
//
// requestIDKey 表示客户端提交的 X-Request-ID 报头名，如果为空则采用 "X-Request-ID"；
func NewContextBuilder(s Server, codec *Codec, requestIDKey string) *ContextBuilder {
	if s == nil {
		panic("s 不能为空")
	}

	if requestIDKey == "" {
		requestIDKey = header.RequestIDKey
	}

	return &ContextBuilder{
		server:       s,
		requestIDKey: requestIDKey,
		codec:        codec,
	}
}

// NewContext 将 w 和 r 包装为 [Context] 对象
//
// 如果出错，则会向 w 输出状态码并返回 nil。
func (b *ContextBuilder) NewContext(w http.ResponseWriter, r *http.Request, route types.Route) *Context {
	id := r.Header.Get(b.requestIDKey)

	h := r.Header.Get(header.Accept)
	mt := b.codec.accept(h)
	if mt == nil {
		var l logs.Recorder = b.server.Logs().DEBUG()
		if id != "" {
			l = b.server.Logs().DEBUG().With(b.requestIDKey, id)
		}
		l.LocaleString(Phrase("not found serialization for %s", h))

		w.WriteHeader(http.StatusNotAcceptable) // 此时 [Context] 未初始化，无法处理 [Problem]，只是简单地输出状态码。
		return nil
	}

	h = r.Header.Get(header.AcceptCharset)
	outputCharsetName, outputCharset := header.ParseAcceptCharset(h)
	if outputCharsetName == "" {
		var l logs.Recorder = b.server.Logs().DEBUG()
		if id != "" {
			l = b.server.Logs().DEBUG().With(b.requestIDKey, id)
		}
		l.LocaleString(Phrase("not found charset for %s", h))

		w.WriteHeader(http.StatusNotAcceptable)
		return nil
	}

	var outputCompressor Compressor
	if b.server.CanCompress() {
		h = r.Header.Get(header.AcceptEncoding)
		var notAcceptable bool
		outputCompressor, notAcceptable = b.codec.acceptEncoding(mt.name(false), h)
		if notAcceptable {
			w.WriteHeader(http.StatusNotAcceptable)
			return nil
		}
	}

	tag := acceptLanguage(b.server, r.Header.Get(header.AcceptLang))

	var inputMimetype UnmarshalFunc
	var inputCharset encoding.Encoding
	h = r.Header.Get(header.ContentType)
	if h != "" {
		var err error
		inputMimetype, inputCharset, err = b.codec.contentType(h)
		if err != nil {
			b.server.Logs().DEBUG().With(b.requestIDKey, id).Error(err)
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return nil
		}
	}

	// 以上是获取构建 Context 的必要参数，并未真正构建 Context，
	// 保证 ID 不为空无任何意义，因为没有后续的执行链可追踪的。
	// 此处开始才会构建 Context 对象，才须确保 ID 不为空。
	if id == "" {
		id = b.server.UniqueID()
	}
	w.Header().Set(b.requestIDKey, id)
	r.Header.Set(b.requestIDKey, id)

	// NOTE: ctx 是从对象池中获取的，所有变量都必须初始化。

	ctx := contextPool.Get().(*Context)
	ctx.b = b
	ctx.route = route
	ctx.request = r
	ctx.exits = ctx.exits[:0]
	ctx.id = id
	ctx.begin = b.server.Now()
	ctx.queries = nil

	// response
	ctx.originResponse = w
	ctx.writer = w
	ctx.outputCompressor = outputCompressor
	ctx.outputCharset = outputCharset
	ctx.outputCharsetName = outputCharsetName
	ctx.status = 0
	ctx.wrote = false
	if ctx.outputCompressor != nil {
		h := ctx.Header()
		h.Set(header.ContentEncoding, outputCompressor.Name())
	}

	ctx.outputMimetype = mt
	ctx.inputMimetype = inputMimetype
	ctx.inputCharset = inputCharset
	ctx.languageTag = tag
	ctx.localePrinter = b.server.NewLocalePrinter(tag)
	ctx.vars = map[any]any{} // TODO: go1.21 可以改为 clear(ctx.vars)
	ctx.logs = b.server.Logs().New(map[string]any{b.requestIDKey: id})

	return ctx
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

	item := ctx.b.codec.accept(mimetype)
	if item == nil {
		panic(fmt.Sprintf("指定的编码 %s 不存在", mimetype))
	}
	ctx.outputMimetype = item
}

// Mimetype 返回输出编码名称
//
// problem 表示是否返回 problem 状态时的值。
func (ctx *Context) Mimetype(problem bool) string {
	if ctx.outputMimetype == nil {
		return ""
	}
	return ctx.outputMimetype.name(problem)
}

// SetEncoding 设置输出的压缩编码
func (ctx *Context) SetEncoding(enc string) {
	if ctx.Wrote() {
		panic("已有内容输出，不可再更改！")
	}
	if ctx.Encoding() == enc {
		return
	}

	compressor, notAcceptable := ctx.b.codec.acceptEncoding(ctx.Mimetype(false), enc)
	if notAcceptable {
		panic(fmt.Sprintf("指定的压缩编码 %s 不存在", enc))
	}
	ctx.outputCompressor = compressor

	if ctx.outputCompressor != nil {
		h := ctx.Header()
		h.Set(header.ContentEncoding, compressor.Name())
	}
}

// Encoding 输出的压缩编码名称
func (ctx *Context) Encoding() string {
	if ctx.outputCompressor == nil {
		return ""
	}
	return ctx.outputCompressor.Name()
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

func (ctx *Context) Free() {
	sliceutil.Reverse(ctx.exits) // TODO: go1.21 改为标准库
	for _, exit := range ctx.exits {
		exit(ctx, ctx.status)
	}
	ctx.logs.Free()

	contextPool.Put(ctx)
}

// OnExit 注册退出当前请求时的处理函数
//
// f 为执行的方法，其原型为：
//
//	func(ctx *Context, status int)
//
// 其中 ctx 即为当前实例，status 则表示实际输出的状态码。
//
// NOTE: 以注册的相反顺序调用这些方法。
func (ctx *Context) OnExit(f func(*Context, int)) { ctx.exits = append(ctx.exits, f) }

func acceptLanguage(s Server, header string) language.Tag {
	if header == "" {
		return s.Language()
	}
	tag, _ := language.MatchStrings(s.Catalog().Matcher(), header)
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
	return strings.ToLower(ctx.Request().Header.Get(header.RequestWithKey)) == header.XHR
}

// Unwrap [http.ResponseController] 通过此方法返回底层的 [http.ResponseWriter]
func (ctx *Context) Unwrap() http.ResponseWriter { return ctx.originResponse }

// Server 获取关联的 Server 实例
func (ctx *Context) Server() Server { return ctx.b.server }
