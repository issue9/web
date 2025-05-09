// SPDX-FileCopyrightText: 2018-2025 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/issue9/logs/v7"
	"github.com/issue9/mux/v9/header"
	"github.com/issue9/mux/v9/types"
	"golang.org/x/text/encoding"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/transform"

	"github.com/issue9/web/compressor"
	"github.com/issue9/web/internal/qheader"
)

var errExitContext = NewLocaleError("exit context")

var contextPool = &sync.Pool{
	New: func() any {
		return &Context{
			exits: make([]OnExitContextFunc, 0, 5),
			vars:  map[any]any{},
		}
	},
}

// ErrExitContext 当已经退出 [Context] 对象时还对其进行操作将返回此错误
func ErrExitContext() error { return errExitContext }

// Context 根据当次 HTTP 请求生成的上下文内容
//
// Context 同时也实现了 [http.ResponseWriter] 接口，
// 但是不推荐非必要情况下直接使用 [http.ResponseWriter] 的接口方法，
// 而是采用返回 [Responser] 的方式向客户端输出内容。
type Context struct {
	s       *InternalServer
	route   types.Route
	request *http.Request
	exits   []OnExitContextFunc
	id      string
	begin   time.Time
	queries *Queries

	done    chan struct{}
	doneErr error

	originResponse    http.ResponseWriter // 原始的 http.ResponseWriter
	writer            io.Writer
	outputCompressor  compressor.Compressor
	outputCharset     encoding.Encoding
	outputCharsetName string
	outputMimetype    *mediaType
	status            int // WriteHeader 保存的副本
	wrote             bool

	// 从客户端提交的 Content-Type 报头解析到的内容
	inputMimetype UnmarshalFunc
	requestBody   io.Reader

	// 区域和本地相关信息
	languageTag   language.Tag
	localePrinter *message.Printer

	// 保存 Context 在存续期间的可复用变量。
	// 这是比 [context.Value] 更经济的变量传递方式，但是这并不是协程安全的。
	vars map[any]any

	logs *AttrLogs
}

// NewContext 将 w 和 r 包装为 [Context] 对象
//
// 如果出错，则会向 w 输出状态码并返回 nil。
func (s *InternalServer) NewContext(w http.ResponseWriter, r *http.Request, route types.Route) *Context {
	id := r.Header.Get(s.requestIDKey)

	debug := func() logs.Recorder { // 根据 id 是否为空返回不同的日志对象
		if id != "" {
			return s.server.Logs().DEBUG().With(s.requestIDKey, id)
		}
		return s.server.Logs().DEBUG()
	}

	h := r.Header.Get(header.Accept)
	mt := s.codec.accept(h) // 空值相当于 */*，如果正确设置，肯定不会返回 nil。
	if mt == nil {
		debug().LocaleString(Phrase("not found serialization for %s", h))
		w.WriteHeader(http.StatusNotAcceptable) // 此时还不知道将 problem 序列化成什么类型，只简单地返回状态码。
		return nil
	}

	h = r.Header.Get(header.AcceptCharset)
	outputCharsetName, outputCharset := qheader.ParseAcceptCharset(h)
	if outputCharsetName == "" {
		debug().LocaleString(Phrase("not found charset for %s", h))
		w.WriteHeader(http.StatusNotAcceptable) // 无法找到对方要求的字符集，依然只是简单地返回状态码。
		return nil
	}

	var outputCompressor compressor.Compressor
	if s.server.CanCompress() {
		var na bool
		if outputCompressor, na = s.codec.acceptEncoding(mt.name(false), r.Header.Get(header.AcceptEncoding)); na {
			w.WriteHeader(http.StatusNotAcceptable)
			return nil
		}
	}

	var inputReader io.Reader = r.Body // 作为服务端使用，Body 始终不为空，且不需要调用 Close，所以类型为 io.Reader 就可以了。
	var inputMimetype UnmarshalFunc
	if h = r.Header.Get(header.ContentType); h != "" {
		var err error
		var inputCharset encoding.Encoding
		if inputMimetype, inputCharset, err = s.codec.contentType(h); err != nil {
			debug().Error(err)
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return nil
		}
		if !qheader.CharsetIsNop(inputCharset) {
			inputReader = transform.NewReader(inputReader, inputCharset.NewDecoder())
		}
	}

	tag := acceptLanguage(s.server, r.Header.Get(header.AcceptLanguage))

	// 以上是获取构建 Context 的必要参数，并未真正构建 Context，
	// 保证 ID 不为空无任何意义，因为没有后续的执行链可追踪的。
	// 此处开始才会构建 Context 对象，须确保 ID 不为空。
	if id == "" {
		id = s.server.UniqueID()
		r.Header.Set(s.requestIDKey, id) // id 本身从 r.Header 获取，所以在 id 不为空的情况下，无须再设置。
	}
	w.Header().Set(s.requestIDKey, id)

	// NOTE: ctx 是从对象池中获取的，所有变量都必须初始化。

	ctx := contextPool.Get().(*Context)
	ctx.s = s
	ctx.route = route
	ctx.request = r
	ctx.exits = ctx.exits[:0]
	ctx.id = id
	ctx.begin = s.server.Now()
	ctx.queries = nil

	ctx.done = make(chan struct{})
	ctx.doneErr = nil

	ctx.originResponse = w
	ctx.writer = w
	ctx.outputCompressor = outputCompressor
	ctx.outputCharset = outputCharset
	ctx.outputCharsetName = outputCharsetName
	ctx.outputMimetype = mt
	ctx.status = 0
	ctx.wrote = false
	if ctx.outputCompressor != nil {
		ctx.Header().Set(header.ContentEncoding, ctx.outputCompressor.Name())
	}

	ctx.inputMimetype = inputMimetype
	ctx.requestBody = inputReader
	ctx.languageTag = tag
	ctx.localePrinter = s.server.Locale().NewPrinter(tag)
	clear(ctx.vars)
	ctx.logs = s.server.Logs().New(map[string]any{s.requestIDKey: id})

	return ctx
}

// GetVar 获取变量
func (ctx *Context) GetVar(key any) (any, bool) {
	v, found := ctx.vars[key]
	return v, found
}

// SetVar 设置变量
func (ctx *Context) SetVar(key, val any) { ctx.vars[key] = val }

// DelVar 删除变量
func (ctx *Context) DelVar(key any) { delete(ctx.vars, key) }

// Route 关联的路由信息
func (ctx *Context) Route() types.Route { return ctx.route }

// Begin 当前对象的初始化时间
func (ctx *Context) Begin() time.Time { return ctx.begin }

// Location 当前用户的时区
//
// 默认情况下，该值与 [Server.Location] 相同。
func (ctx *Context) Location() *time.Location { return ctx.Begin().Location() }

// Now 相当于指定了时区的 [time.Now]
func (ctx *Context) Now() time.Time { return time.Now().In(ctx.Location()) }

// SetLocation 改变当前用户的时区
//
// 该操作也将同时改变 [Context.Begin] 的时区。
func (ctx *Context) SetLocation(l *time.Location) { ctx.begin = ctx.Begin().In(l) }

func (ctx *Context) ParseTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, ctx.Location())
}

// ID 当前请求的唯一 ID
//
// 一般源自客户端的 X-Request-ID 报头，如果不存在，则由 [Server.UniqueID] 生成。
func (ctx *Context) ID() string { return ctx.id }

// SetCharset 设置输出的字符集
//
// 不会修改 [Context.Request] 中的 Accept-Charset 报头，如果需要同时修改此报头
// 可以通过 [Server.NewContext] 构建一个新 [Context] 对象。
func (ctx *Context) SetCharset(charset string) {
	if ctx.Wrote() {
		panic("已有内容输出，不可再更改！")
	}
	if ctx.Charset() == charset {
		return
	}

	outputCharsetName, outputCharset := qheader.ParseAcceptCharset(charset)
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
// 不会修改 [Context.Request] 中的 Accept 报头，如果需要同时修改此报头
// 可以通过 [Server.NewContext] 构建一个新 [Context] 对象。
func (ctx *Context) SetMimetype(mimetype string) {
	if ctx.Wrote() {
		panic("已有内容输出，不可再更改！")
	}
	if ctx.Mimetype(false) == mimetype {
		return
	}

	item := ctx.s.codec.accept(mimetype)
	if item == nil {
		panic(fmt.Sprintf("指定的编码 %s 不存在", mimetype))
	}
	ctx.outputMimetype = item
}

// Mimetype 返回输出编码名称
//
// problem 表示是否返回 problem 状态时的值。
func (ctx *Context) Mimetype(problem bool) string { return ctx.outputMimetype.name(problem) }

// SetEncoding 设置输出的压缩编码
//
// 不会修改 [Context.Request] 中的 Accept-Encoding 报头，如果需要同时修改此报头
// 可以通过 [Server.NewContext] 构建一个新 [Context] 对象。
func (ctx *Context) SetEncoding(enc string) {
	if ctx.Wrote() {
		panic("已有内容输出，不可再更改！")
	}
	if ctx.Encoding() == enc {
		return
	}

	c, notAcceptable := ctx.s.codec.acceptEncoding(ctx.Mimetype(false), enc)
	if notAcceptable {
		panic(fmt.Sprintf("指定的压缩编码 %s 不存在", enc))
	}
	ctx.outputCompressor = c

	if ctx.outputCompressor != nil {
		h := ctx.Header()
		h.Set(header.ContentEncoding, c.Name())
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
// 不会修改 [Context.Request] 中的 Accept-Language 报头，如果需要同时修改此报头
// 可以通过 [Server.NewContext] 构建一个新 [Context] 对象。
func (ctx *Context) SetLanguage(tag language.Tag) {
	// 不判断是否有内容已经输出，允许中途改变语言。
	if ctx.languageTag != tag {
		ctx.languageTag = tag
		ctx.localePrinter = ctx.Server().Locale().NewPrinter(tag)
	}
}

func (ctx *Context) LocalePrinter() *message.Printer { return ctx.localePrinter }

func (ctx *Context) LanguageTag() language.Tag { return ctx.languageTag }

func (s *InternalServer) freeContext(ctx *Context) {
	for _, exit := range ctx.exits {
		exit(ctx, ctx.status)
	}

	for _, f := range s.exitContexts {
		f(ctx, ctx.status)
	}

	// 以下开始回收内存

	close(ctx.done)
	ctx.doneErr = ErrExitContext()

	logs.FreeAttrLogs(ctx.logs)
	contextPool.Put(ctx)
}

// OnExit 注册退出当前请求时的处理函数
//
// 此方法添加的函数会先于 [Server.OnExitContext] 添加的函数执行。
func (ctx *Context) OnExit(f OnExitContextFunc) { ctx.exits = append(ctx.exits, f) }

func acceptLanguage(s Server, h string) language.Tag {
	if h == "" {
		return s.Locale().ID()
	}
	tag, _ := language.MatchStrings(s.Locale().Matcher(), h)
	return tag
}

// ClientIP 返回客户端的 IP 地址及端口
//
// 获取顺序如下：
//   - X-Forwarded-For 的第一个元素
//   - Remote-Addr 报头
//   - X-Read-IP 报头
func (ctx *Context) ClientIP() string { return qheader.ClientIP(ctx.Request()) }

// Logs 返回日志操作对象
//
// 当前返回实例的日志输出时会带上当前请求的 [Context.ID] 作为额外参数。
func (ctx *Context) Logs() *AttrLogs { return ctx.logs }

func (ctx *Context) IsXHR() bool {
	return strings.ToLower(ctx.Request().Header.Get(header.XRequestedWith)) == "xmlhttprequest"
}

var idempotentMethods = []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions, http.MethodTrace}

// Idempotent 请求方法是否为[幂等]
//
// [幂等]: https://developer.mozilla.org/zh-CN/docs/Glossary/Idempotent
func (ctx *Context) Idempotent() bool {
	return slices.Index(idempotentMethods, ctx.Request().Method) >= 0
}

// Unwrap [http.ResponseController] 通过此方法返回底层的 [http.ResponseWriter]
func (ctx *Context) Unwrap() http.ResponseWriter { return ctx.originResponse }

// Server 获取关联的 [Server] 实例
func (ctx *Context) Server() Server { return ctx.s.server }

// Deadline 接口 [context.Context] 的方法。
//
// 未提供实际的功能
func (ctx *Context) Deadline() (time.Time, bool) { return time.Time{}, false }

// Value 接口 [context.Context] 的方法
//
// 功能与 [Context.GetVar] 是相同的。
func (ctx *Context) Value(key any) any {
	val, _ := ctx.GetVar(key)
	return val
}

// Done 接口 [context.Context] 的方法
func (ctx *Context) Done() <-chan struct{} { return ctx.done }

// Err 接口 [context.Context] 的方法
func (ctx *Context) Err() error { return ctx.doneErr }
