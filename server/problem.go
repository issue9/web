// SPDX-License-Identifier: MIT

package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v4"
	"github.com/issue9/sliceutil"
	"golang.org/x/text/message"
)

const aboutBlank = "about:blank"

type (
	// Problem API 错误信息对象需要实现的接口
	//
	// Problem 是对 [Responser] 细化，用于反馈给用户非正常状态下的数据，
	// 比如用户提交的数据错误，往往会返回 400 的状态码，
	// 并附带一些具体的字段错误信息，此类数据都可以以 Problem 对象的方式反馈给用户。
	//
	// 除了当前接口，该对象可能还要实现相应的序列化接口，比如要能被 JSON 解析，
	// 就要实现 json.Marshaler 接口或是相应的 struct tag。
	//
	// 并未规定实现者输出的字段名和布局，实现者可以根据 [BuildProblemFunc]
	// 给定的参数，结合自身需求决定。比如 [RFC7807Builder] 是对 [RFC7807] 的实现。
	//
	// [RFC7807]: https://datatracker.ietf.org/doc/html/rfc7807
	Problem interface {
		Responser

		// With 添加新的输出字段
		//
		// 如果添加的字段名称与现有的字段重名，应当 panic。
		With(key string, val any) Problem

		// AddParam 添加数据验证错误信息
		AddParam(name string, reason string) Problem
	}

	// BuildProblemFunc 生成 [Problem] 对象的方法
	//
	// id 表示当前错误信息的唯一值，这将是一个标准的 URL，指向线上的文档地址，有可能不会真实存在；
	// title 错误信息的简要描述；
	// status 输出的状态码；
	BuildProblemFunc func(id, title string, status int) Problem

	Problems struct {
		builder   BuildProblemFunc
		baseURL   string
		blank     bool             // 不输出 id 值
		problems  []*statusProblem // 不用 map，保证 Visit 以同样的顺序输出。
		mimetypes map[string]string
	}

	statusProblem struct {
		id     string
		status int
		title  localeutil.LocaleStringer
		detail localeutil.LocaleStringer
	}

	// CTXSanitizer 在 [Context] 关联的上下文环境中对数据进行验证和修正
	//
	// 在 [Context.Read]、[Context.QueryObject] 以及 [Queries.Object]
	// 中会在解析数据成功之后会调用该接口。
	CTXSanitizer interface {
		// CTXSanitize 验证和修正当前对象的数据
		CTXSanitize(*Validation)
	}
)

func (srv *Server) Problems() *Problems { return srv.problems }

func newProblems(f BuildProblemFunc) *Problems {
	p := &Problems{
		builder:   f,
		problems:  make([]*statusProblem, 0, 50),
		mimetypes: make(map[string]string, 10),
	}

	add := func(s int) {
		msg := localeutil.Phrase(http.StatusText(s))
		p.Add(strconv.Itoa(s), s, msg, msg)
	}

	// 400
	add(http.StatusBadRequest)
	add(http.StatusUnauthorized)
	add(http.StatusPaymentRequired)
	add(http.StatusForbidden)
	add(http.StatusNotFound)
	add(http.StatusMethodNotAllowed)
	add(http.StatusNotAcceptable)
	add(http.StatusProxyAuthRequired)
	add(http.StatusRequestTimeout)
	add(http.StatusConflict)
	add(http.StatusGone)
	add(http.StatusLengthRequired)
	add(http.StatusPreconditionFailed)
	add(http.StatusRequestEntityTooLarge)
	add(http.StatusRequestURITooLong)
	add(http.StatusUnsupportedMediaType)
	add(http.StatusRequestedRangeNotSatisfiable)
	add(http.StatusExpectationFailed)
	add(http.StatusTeapot)
	add(http.StatusMisdirectedRequest)
	add(http.StatusUnprocessableEntity)
	add(http.StatusLocked)
	add(http.StatusFailedDependency)
	add(http.StatusTooEarly)
	add(http.StatusUpgradeRequired)
	add(http.StatusPreconditionRequired)
	add(http.StatusTooManyRequests)
	add(http.StatusRequestHeaderFieldsTooLarge)
	add(http.StatusUnavailableForLegalReasons)

	// 500
	add(http.StatusInternalServerError)
	add(http.StatusNotImplemented)
	add(http.StatusBadGateway)
	add(http.StatusServiceUnavailable)
	add(http.StatusGatewayTimeout)
	add(http.StatusHTTPVersionNotSupported)
	add(http.StatusVariantAlsoNegotiates)
	add(http.StatusInsufficientStorage)
	add(http.StatusLoopDetected)
	add(http.StatusNotExtended)
	add(http.StatusNetworkAuthenticationRequired)

	return p
}

// BaseURL [BuildProblemFunc] 参数 id 的前缀
//
// 返回的内容说明，可参考 [Problems.SetBaseURL]。
func (p *Problems) BaseURL() string { return p.baseURL }

// SetBaseURL 设置传递给 [BuildProblemFunc] 中 id 参数的前缀
//
// [Problem] 实现者可以根据自由决定 id 字终以什么形式展示，
// 此处的设置只是决定了传递给 [BuildProblemFunc] 的 id 参数格式。
// 可以有以下三种形式：
//
//   - 空值 不作任何改变；
//   - about:blank 将传递空值给 [BuildProblemFunc]；
//   - 其它非空值 以前缀形式附加在原本的 id 之上；
func (p *Problems) SetBaseURL(base string) {
	p.baseURL = base
	p.blank = base == aboutBlank
}

// Add 添加新的错误类型
//
// id 表示该错误的唯一值，如果存在相同会 panic；
// [Problems.Problem] 会根据此值查找相应的文字说明给予 title 字段；
// status 表示输出给客户端的状态码；
// title 和 detail 表示此 id 关联的简要说明和详细说明；
func (p *Problems) Add(id string, status int, title, detail localeutil.LocaleStringer) *Problems {
	if p.Exists(id) {
		panic(fmt.Sprintf("存在相同值的 id 参数 %s", id))
	}

	sp := &statusProblem{status: status, title: title, detail: detail, id: id}
	p.problems = append(p.problems, sp)
	return p
}

func (p *Problems) Count() int { return len(p.problems) }

func (p *Problems) Exists(id string) bool {
	return sliceutil.Exists(p.problems, func(sp *statusProblem) bool { return sp.id == id })
}

// AddMimetype 指定返回 [Problem] 时的 content-type 值
//
// mimetype 为正常情况下的 content-type 值，当输出对象为 [Problem] 时，
// 可以指定不同的值，比如 application/json 可以对应输出 application/problem+json，
// 这也是 RFC7807 推荐的作法。
func (p *Problems) AddMimetype(mimetype, problemType string) *Problems {
	if _, exists := p.mimetypes[mimetype]; exists {
		panic(fmt.Sprintf("已经存在的 mimetype %s", mimetype))
	}
	p.mimetypes[mimetype] = problemType
	return p
}

func (p *Problems) mimetype(mimetype string) string {
	if v, exists := p.mimetypes[mimetype]; exists {
		return v
	}
	return mimetype
}

// Visit 遍历所有由 [Problems.Add] 添加的项
//
// f 为遍历的函数，其原型为：
//
//	func(id string, status int, title, detail localeutil.LocaleStringer)
//
// 分别对应 [Problems.Add] 添加时的各个参数。
//
// 用户可以通过此方法生成 QA 页面。
func (p *Problems) Visit(f func(string, int, localeutil.LocaleStringer, localeutil.LocaleStringer) bool) {
	for _, item := range p.problems {
		if !f(item.id, item.status, item.title, item.detail) {
			return
		}
	}
}

// Problem 根据 id 生成 [Problem] 对象
func (p *Problems) Problem(printer *message.Printer, id string) Problem {
	sp, found := sliceutil.At(p.problems, func(sp *statusProblem) bool { return sp.id == id })
	if !found { // 这更像是编译期错误，所以直接 panic。
		panic(fmt.Sprintf("未找到有关 %s 的定义", id))
	}

	if p.blank {
		id = ""
	} else {
		id = p.baseURL + id
	}
	return p.builder(id, sp.title.LocaleString(printer), sp.status)
}

// Problem 转换成 [Problem] 对象
//
// 如果当前对象没有收集到错误，那么将返回 nil。
func (v *Validation) Problem(id string) Problem {
	if v == nil || v.Count() == 0 {
		return nil
	}

	p := v.Context().Problem(id)
	for index, key := range v.keys {
		p.AddParam(key, v.reasons[index])
	}
	return p
}

// Problem 返回批定 id 的错误信息
//
// id 通过此值从 [Problems] 中查找相应在的 title 并赋值给返回对象；
func (ctx *Context) Problem(id string) Problem {
	return ctx.Server().Problems().Problem(ctx.LocalePrinter(), id)
}

// InternalServerError 输出 ERROR 通道并向返回 500 表示的 [Problem] 对象
func (ctx *Context) InternalServerError(err error) Problem {
	return ctx.logError(3, "500", logs.LevelError, err)
}

// Error 将 err 输出到日志并以指定 id 的 [Problem] 返回
func (ctx *Context) Error(id string, level logs.Level, err error) Problem {
	return ctx.logError(3, id, level, err)
}

func (ctx *Context) logError(depth int, id string, level logs.Level, err error) Problem {
	entry := ctx.Logs().NewEntry(level).Location(depth)
	if le, ok := err.(localeutil.LocaleStringer); ok {
		entry.Message = le.LocaleString(ctx.LocalePrinter())
	} else {
		entry.Message = err.Error()
	}
	ctx.Logs().Output(entry)
	return ctx.Problem(id)
}

// NotFound 返回 id 值为 404 的 [Problem] 对象
func (ctx *Context) NotFound() Problem { return ctx.Problem("404") }

func (ctx *Context) NotImplemented() Problem { return ctx.Problem("501") }
