// SPDX-License-Identifier: MIT

package server

import (
	"fmt"

	"github.com/issue9/localeutil"
)

const aboutBlank = "about:blank"

// Problem API 错误信息对象需要实现的接口
//
// 除了当前接口，该对象可能还要实现相应的序列化接口，比如要能被 JSON 解析，
// 就要实现 json.Marshaler 接口或是相应的 struct tag。
//
// 并未规定 [Problem] 实现都输出的字段以及布局，实现者可以根据 [BuildProblemFunc]
// 给定的参数，结合自身需求决定。比如 [RFC7807Builder] 实现了一个简要的
// RFC7807 标准的错误信息对象。
type Problem interface {
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
// id 表示当前错误信息的唯一值，这将是一个标准的 URL，指向线上的文档地址；
// title 错误信息的简要描述；
// status 输出的状态码，该值将由 [Problem.Status] 返回；
type BuildProblemFunc func(id string, title localeutil.LocaleStringer, status int) Problem

type Problems struct {
	builder   BuildProblemFunc
	baseURL   string
	blank     bool // Problems.Problem 不输出 id 值
	problems  map[string]*statusProblem
	mimetypes map[string]string
}

type statusProblem struct {
	status int
	title  localeutil.LocaleStringer
	detail localeutil.LocaleStringer
}

func newProblems(f BuildProblemFunc) *Problems {
	return &Problems{
		builder:   f,
		problems:  make(map[string]*statusProblem, 50),
		mimetypes: make(map[string]string, 10),
	}
}

// BaseURL [BuildFunc] 参数 id 的前缀
//
// 返回的内容说明，可参考 [Problems.SetBaseURL]。
func (p *Problems) BaseURL() string { return p.baseURL }

// SetBaseURL 设置传递给 [BuildFunc] 中 id 参数的前缀
//
// [Problem] 实现者可以根据自由决定 id 字终以什么形式展示，
// 此处的设置只是决定了传递给 [BuildFunc] 的 id 参数格式。
// 可以有以下三种形式：
//
//   - 空值 不作任何改变；
//   - about:blank 将传递空值给 [BuildFunc]；
//   - 其它非空值 以前缀形式附加在原本的 id 之上；
func (p *Problems) SetBaseURL(base string) {
	p.baseURL = base
	p.blank = base == aboutBlank
}

// Add 添加新的错误类型
//
// id 表示该错误的唯一值；
// [Problems.Problem] 会根据此值查找相应的文字说明给予 title 字段；
// status 表示输出给客户端的状态码；
// title 和 detail 表示此 id 关联的简要说明和详细说明。title 会出现在 [Problems.Problem] 返回的对象中。
func (p *Problems) Add(id string, status int, title, detail localeutil.LocaleStringer) {
	if _, found := p.problems[id]; found {
		panic("存在相同值的 id 参数")
	}
	p.problems[id] = &statusProblem{status: status, title: title, detail: detail}
}

// AddMimetype 添加输出的 mimetype 值
//
// mimetype 为正常情况下输出的值，当输出对象为 [Problem] 时，可以指定一个特殊的值，
// 比如 application/json 可以对应输出 application/problem+json，
// 这也是 RFC7807 推荐的作法。
func (p *Problems) AddMimetype(mimetype, problemType string) {
	if _, exists := p.mimetypes[mimetype]; exists {
		panic("已经存在的 mimetype")
	}
	p.mimetypes[mimetype] = problemType
}

func (p *Problems) mimetype(mimetype string) string {
	if v, exists := p.mimetypes[mimetype]; exists {
		return v
	}
	return mimetype
}

// Visit 遍历所有 Add 添加的项
//
// f 为遍历的函数，其原型为：
//
//	func(id string, status int, title, detail localeutil.LocaleStringer)
//
// 分别对应 [Problems.Add] 添加时的各个参数。
//
// 用户可以通过此方法生成 QA 页面。
func (p *Problems) Visit(f func(string, int, localeutil.LocaleStringer, localeutil.LocaleStringer) bool) {
	for t, item := range p.problems {
		if !f(t, item.status, item.title, item.detail) {
			return
		}
	}
}

// Problem 根据 id 生成 [Problem] 对象
//
// id 通过此值查找相应的 title；
func (p *Problems) Problem(id string) Problem {
	sp, found := p.problems[id]
	if !found {
		panic(fmt.Sprintf("未找到有关 %s 的定义", id))
	}

	if p.blank {
		id = ""
	} else {
		id = p.baseURL + id
	}
	return p.builder(id, sp.title, sp.status)
}
