// SPDX-License-Identifier: MIT

package problem

import (
	"fmt"

	"github.com/issue9/localeutil"
	"golang.org/x/text/message"
)

const aboutBlank = "about:blank"

type Problems struct {
	builder  BuildFunc
	baseURL  string
	problems map[string]*statusProblem
	blank    bool // Problems.Problem 不输出 id 值
}

type statusProblem struct {
	status int
	title  localeutil.LocaleStringer
	detail localeutil.LocaleStringer
}

func NewProblems(f BuildFunc) *Problems {
	return &Problems{
		builder:  f,
		problems: make(map[string]*statusProblem, 50),
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
//  - 空值 不作任何改变；
//  - about:blank 将传递空值给 [BuildFunc]；
//  - 其它非空值 以前缀形式附加在原本的 id 之上；
func (p *Problems) SetBaseURL(base string) {
	p.baseURL = base
	p.blank = base == aboutBlank
}

// AddProblem 添加新的错误类型
//
// id 表示该错误的唯一值；
// [Problems.Problem] 会根据此值查找相应的文字说明给予 title 字段；
// status 表示输出给客户端的状态码；
// title 和 detail 表示此 id 关联的简要说明和详细说明。title 会出现在 [Problems.Problem] 返回的对象中。
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
// id 通过此值查找相应的 title；
func (p *Problems) Problem(id string, printer *message.Printer) Problem {
	sp, found := p.problems[id]
	if !found {
		panic(fmt.Sprintf("未找到有关 %s 的定义", id))
	}

	if p.blank {
		id = ""
	} else {
		id = p.baseURL + id
	}
	return p.builder(id, sp.title.LocaleString(printer), sp.status)
}

// Problem 将验证结果转换成 [Problem] 对象
//
// 转换成 Problem 对象之后，v 随之将被释放。
// 如果 v.Count() == 0，那么将返回 nil。
func (v *Validation) Problem(ps *Problems, id string, p *message.Printer) Problem {
	if v.Count() > 0 {
		pp := ps.Problem(id, p)
		v.Visit(func(key string, reason localeutil.LocaleStringer) bool {
			pp.AddParam(key, reason.LocaleString(p))
			return true
		})
		v.Destroy()
		return pp
	}
	v.Destroy()
	return nil
}
