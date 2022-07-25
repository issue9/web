// SPDX-License-Identifier: MIT

package problem

import (
	"fmt"

	"github.com/issue9/localeutil"
	"golang.org/x/text/message"
)

type Problems struct {
	builder     BuildFunc
	typeBaseURL string
	problems    map[string]*statusProblem
	blank       bool // 在输出时所有的 Type 强制为 about:blank
}

type statusProblem struct {
	status int
	title  localeutil.LocaleStringer
	detail localeutil.LocaleStringer
}

// NewProblems 声明 [Problems] 对象
//
// base type 字段的基地址；
// blank [Problems.Problem] 生成的对象是否包含 id。
func NewProblems(f BuildFunc, base string, blank bool) *Problems {
	return &Problems{
		builder:     f,
		typeBaseURL: base,
		blank:       blank,
		problems:    make(map[string]*statusProblem, 50),
	}
}

// AddProblem 添加新的错误类型
//
// id 表示该错误的唯一值。
// [Problems.Problem] 会根据此值查找相应的文字说明给予 title 和 detail 字段；
// status 表示输出给客户端的状态码；
// title 和 detail 表示此 id 关联的简要说明和详细说明，
// 这些值有可能会赋予实现 [Problem] 接口的对象；
func (p *Problems) Add(id string, status int, title, detail localeutil.LocaleStringer) {
	if p.problems == nil {
	}

	if _, found := p.problems[id]; found {
		panic("存在相同值的 id 参数")
	}
	p.problems[id] = &statusProblem{status: status, title: title, detail: detail}
}

// Visit 遍历所有 Add 添加的项
//
// f 为遍历的函数，其原型为：
//  func(id string, status int, title, detail localeutil.LocaleStringer)
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
// id 通过此值查找相应的 title 和 detail 值；
func (p *Problems) Problem(id string, printer *message.Printer) Problem {
	sp, found := p.problems[id]
	if !found {
		panic(fmt.Sprintf("未找到有关 %s 的定义", id))
	}

	if p.blank {
		id = ""
	} else {
		id = p.typeBaseURL + id
	}
	return p.builder(id, sp.title.LocaleString(printer), sp.detail.LocaleString(printer), sp.status)
}
