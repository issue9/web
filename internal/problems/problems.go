// SPDX-License-Identifier: MIT

// Package problems 提供对 Problem 相关内容的管理
package problems

import (
	"fmt"

	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"
)

// Problems 管理 Problem
//
// P 表示 Problem 接口类型
type Problems[P any] struct {
	builder  func(id string, status int, title, detail string) P
	prefix   string
	problems []*statusProblem // 不用 map，保证 Problems 顺序相同。
}

type statusProblem struct {
	id string // 带 prefix
	ID string // 用户指定的原始值

	Status int
	Title  localeutil.LocaleStringer
	Detail localeutil.LocaleStringer
}

func New[P any](prefix string, builder func(id string, status int, title, detail string) P) *Problems[P] {
	p := &Problems[P]{
		builder:  builder,
		prefix:   prefix,
		problems: make([]*statusProblem, 0, 50),
	}

	for id, status := range statuses {
		text := "problem." + id
		title := localeutil.Phrase(text)
		detail := localeutil.Phrase(text + ".detail")
		p.Add(id, status, title, detail)
	}

	return p
}

func (p *Problems[P]) Add(id string, status int, title, detail localeutil.LocaleStringer) {
	if p.exists(id) {
		panic(fmt.Sprintf("存在相同值的 id 参数 %s", id))
	}
	if !IsValidStatus(status) {
		panic("status 必须是一个有效的状态码")
	}

	if title == nil {
		panic("title 不能为空")
	}

	s := &statusProblem{ID: id, Status: status, Title: title, Detail: detail}
	if p.prefix == ProblemAboutBlank {
		s.id = ProblemAboutBlank
	} else {
		s.id = p.prefix + s.ID
	}
	p.problems = append(p.problems, s)
}

func (p *Problems[P]) exists(id string) bool {
	return sliceutil.Exists(p.problems, func(sp *statusProblem) bool { return sp.ID == id })
}

func (p *Problems[P]) Visit(visit func(prefix, id string, status int, title, detail localeutil.LocaleStringer)) {
	for _, s := range p.problems {
		visit(p.prefix, s.ID, s.Status, s.Title, s.Detail)
	}
}

func (p *Problems[P]) Problem(printer *localeutil.Printer, id string) P {
	sp, found := sliceutil.At(p.problems, func(sp *statusProblem) bool { return sp.ID == id })
	if !found { // 初始化时没有给定相关的定义，所以直接 panic。
		panic(fmt.Sprintf("未找到有关 %s 的定义", id))
	}

	return p.builder(sp.id, sp.Status, sp.Title.LocaleString(printer), sp.Detail.LocaleString(printer))
}

func IsValidStatus(status int) bool { return status >= 100 && status < 600 }
