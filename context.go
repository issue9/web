// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"
	"unique"

	"github.com/issue9/logs/v7"
	"github.com/issue9/mux/v10/header"
	"github.com/issue9/mux/v10/types"
	"golang.org/x/text/encoding"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/transform"

	"github.com/issue9/web/internal/qheader"
	"github.com/issue9/web/problems"
)

var contextPool = &sync.Pool{
	New: func() any {
		return &Context{
			exits: make([]OnExitContextFunc, 0, 5),
			vars:  map[any]any{},
		}
	},
}

// Context 根据当次 HTTP 请求生成的上下文内容
//
// Context 同时也实现了 [http.ResponseWriter] 接口，
// 但是不推荐非必要情况下直接使用 [http.ResponseWriter] 的接口方法，
// 而是采用返回 [Responser] 的方式向客户端输出内容。
type Context struct {
	baseContext

	s       *Server
	route   *types.Route
	request *http.Request
	exits   []OnExitContextFunc
	id      string
	begin   time.Time
	queries *Queries

	originResponse    http.ResponseWriter // 原始的 http.ResponseWriter
	writer            io.Writer
	outputCompressor  Compressor
	outputCharset     encoding.Encoding
	outputCharsetName string
	outputMediaType   MediaType
	status            int // WriteHeader 保存的副本，非零表示已经输出。
	wrote             bool

	// 从客户端提交的 Content-Type 报头解析到的内容
	inputUnmarshal UnmarshalFunc
	requestBody    io.Reader

	// 区域和本地相关信息
	languageTag   language.Tag
	localePrinter *message.Printer

	// 保存 Context 在存续期间的可复用变量。
	// 这是比 [context.Value] 更经济的变量传递方式，但是这并不是协程安全的。
	vars map[any]any

	logs *AttrLogs
}

func parseHeaderErr(header string, err error) LocaleStringer {
	return Phrase("parse http header %[1]s error %[2]s", header, err.Error())
}

// 输出降级的错误页面
// 在 Context 未初始化成功时，可以使用此方法
func (s *Server) errorPage(w http.ResponseWriter, id unique.Handle[string]) {
	item := s.Problems().Get(id)
	txt := item.Title.LocaleString(s.Locale().Printer())
	http.Error(w, txt, item.Status)
}

// NewContext 将 w 和 r 包装为 [Context] 对象
//
// 这适合从标准库的请求中创建 [Context] 对象。
// [types.Route] 类型的参数需要用户通过 [types.NewRoute] 自行创建。
// 如果出错，则会向 w 输出状态码并返回 nil。
//
// NOTE: 由此方法创建的对象在整个会话结束后会被回收。
func (s *Server) NewContext(w http.ResponseWriter, r *http.Request, route *types.Route) *Context {
	id := r.Header.Get(s.requestIDKey)

	warn := func() logs.Recorder { // 根据 id 是否为空返回不同的日志对象
		if id != "" {
			return s.Logs().WARN().With(s.requestIDKey, id)
		}
		return s.Logs().WARN()
	}

	h := r.Header.Get(header.Accept)
	mt, err := s.codec.accept(h) // 空值相当于 */*，如果正确设置，肯定不会返回 nil。
	if err != nil {
		warn().LocaleString(parseHeaderErr(header.Accept, err))
		s.errorPage(w, problems.NotAcceptable)
		return nil
	}
	if mt == nil {
		warn().LocaleString(Phrase("not found serialization for %s", h))
		s.errorPage(w, problems.NotAcceptable)
		return nil
	}

	h = r.Header.Get(header.AcceptCharset)
	outputCharsetName, outputCharset, errs := qheader.ParseAcceptCharset(h)
	if len(errs) > 0 {
		warn().LocaleString(parseHeaderErr(header.AcceptCharset, unableParse(strings.Join(errs, ","))))
		s.errorPage(w, problems.NotAcceptable)
		return nil
	}
	if outputCharsetName == "" {
		warn().LocaleString(Phrase("not found charset for %s", h))
		s.errorPage(w, problems.NotAcceptable)
		return nil
	}

	var outputCompressor Compressor
	if s.codec.CanCompress() {
		if outputCompressor, err = s.codec.acceptEncoding(mt.Name(), r.Header.Get(header.AcceptEncoding)); err != nil {
			warn().LocaleString(parseHeaderErr(header.AcceptEncoding, err))
			s.errorPage(w, problems.NotAcceptable)
			return nil
		}
	}

	var inputReader io.Reader = r.Body // 作为服务端使用，Body 始终不为空，且不需要调用 Close，所以类型为 io.Reader 就可以了。
	var unmarshal UnmarshalFunc
	if h = r.Header.Get(header.ContentType); h != "" {
		var err error
		var inputCharset encoding.Encoding
		if unmarshal, inputCharset, err = s.codec.contentType(h); err != nil {
			warn().Error(err)
			s.errorPage(w, problems.UnsupportedMediaType)
			return nil
		}
		if !qheader.CharsetIsNop(inputCharset) {
			inputReader = transform.NewReader(inputReader, inputCharset.NewDecoder())
		}
	}

	tag := acceptLanguage(s, r.Header.Get(header.AcceptLanguage))

	// 以上是获取构建 Context 的必要参数，并未真正构建 Context，
	// 保证 ID 不为空无任何意义，因为没有后续的执行链可追踪的。
	// 此处开始才会构建 Context 对象，须确保 ID 不为空。
	if id == "" {
		id = s.UniqueID()
		r.Header.Set(s.requestIDKey, id) // id 本身从 r.Header 获取，所以在 id 不为空的情况下，无须再设置。
	}
	w.Header().Set(s.requestIDKey, id)

	c, cancel := context.WithCancel(r.Context())

	// NOTE: ctx 是从对象池中获取的，所有变量都必须初始化。

	ctx := contextPool.Get().(*Context)
	ctx.Context = c
	ctx.cancel = cancel

	ctx.s = s
	ctx.route = route
	ctx.request = r
	ctx.exits = ctx.exits[:0]
	ctx.id = id
	ctx.begin = s.Now()
	ctx.queries = nil

	ctx.originResponse = w
	ctx.writer = w
	ctx.outputCompressor = outputCompressor
	ctx.outputCharset = outputCharset
	ctx.outputCharsetName = outputCharsetName
	ctx.outputMediaType = mt
	ctx.status = 0
	ctx.wrote = false

	ctx.inputUnmarshal = unmarshal
	ctx.requestBody = inputReader
	ctx.languageTag = tag
	ctx.localePrinter = s.Locale().NewPrinter(tag)
	clear(ctx.vars)
	ctx.logs = s.Logs().New(map[string]any{s.requestIDKey: id})

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
func (ctx *Context) Route() *types.Route { return ctx.route }

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

	outputCharsetName, outputCharset, errs := qheader.ParseAcceptCharset(charset)
	if len(errs) > 0 {
		panic(fmt.Sprintf("无法解析以下内容：%s", strings.Join(errs, ",")))
	}
	if outputCharsetName == "" {
		panic(fmt.Sprintf("指定的字符集 %s 不存在", charset))
	}
	ctx.outputCharset = outputCharset
	ctx.outputCharsetName = outputCharsetName
}

// Charset 输出的字符集
func (ctx *Context) Charset() string { return ctx.outputCharsetName }

// SetMediaType 设置输出的格式
//
// 不会修改 [Context.Request] 中的 Accept 报头，如果需要同时修改此报头
// 可以通过 [Server.NewContext] 构建一个新 [Context] 对象。
func (ctx *Context) SetMediaType(mediaType string) {
	if ctx.Wrote() {
		panic("已有内容输出，不可再更改！")
	}
	if ctx.MediaType(false) == mediaType {
		return
	}

	item, err := ctx.s.codec.accept(mediaType)
	if err != nil {
		panic(fmt.Sprintf("解析参数时发生了错误：%s", err.Error()))
	}
	if item == nil {
		panic(fmt.Sprintf("指定的编码 %s 不存在", mediaType))
	}
	ctx.outputMediaType = item
}

// MediaType 返回输出编码名称
//
// problem 表示是否返回 problem 状态时的值。
func (ctx *Context) MediaType(problem bool) string {
	if problem {
		return ctx.outputMediaType.ProblemName()
	}
	return ctx.outputMediaType.Name()
}

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

	c, err := ctx.s.codec.acceptEncoding(ctx.MediaType(false), enc)
	if err != nil {
		panic(fmt.Sprintf("解析参数时发生了错误：%s", err.Error()))
	}
	ctx.outputCompressor = c
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

func (ctx *Context) freeContext() {
	for _, exit := range ctx.exits {
		exit(ctx, ctx.status)
	}
	clear(ctx.exits) // 释放引用

	for _, f := range ctx.s.exitContexts {
		f(ctx, ctx.status)
	}

	// 以下开始回收内存
	ctx.cancel()

	logs.FreeAttrLogs(ctx.logs)
	contextPool.Put(ctx)
}

// OnExit 注册退出当前请求时的处理函数
//
// 此方法添加的函数会先于 [Server.OnExitContext] 添加的函数执行。
func (ctx *Context) OnExit(f OnExitContextFunc) { ctx.exits = append(ctx.exits, f) }

func acceptLanguage(s *Server, h string) language.Tag {
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
//   - X-Real-IP 报头
func (ctx *Context) ClientIP() string { return qheader.ClientIP(ctx.Request()) }

// Logs 返回日志操作对象
//
// 当前返回实例的日志输出时会带上当前请求的 [Context.ID] 作为额外参数。
func (ctx *Context) Logs() *AttrLogs { return ctx.logs }

func (ctx *Context) IsXHR() bool {
	return strings.ToLower(ctx.Request().Header.Get(header.XRequestedWith)) == "xmlhttprequest"
}

var idempotentMethods = []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodDelete, http.MethodOptions, http.MethodTrace}

// IsIdempotent 请求方法是否为[幂等]
//
// [幂等]: https://developer.mozilla.org/zh-CN/docs/Glossary/IsIdempotent
func (ctx *Context) IsIdempotent() bool {
	return slices.Contains(idempotentMethods, ctx.Request().Method)
}

// Unwrap [http.ResponseController] 通过此方法返回底层的 [http.ResponseWriter]
func (ctx *Context) Unwrap() http.ResponseWriter { return ctx.originResponse }

// Server 获取关联的 [Server] 实例
func (ctx *Context) Server() *Server { return ctx.s }

// Value 接口 [context.Context] 的方法
//
// 功能与 [Context.GetVar] 是相同的。
func (ctx *Context) Value(key any) any {
	if val, found := ctx.GetVar(key); found {
		return val
	}
	return ctx.Context.Value(key)
}
