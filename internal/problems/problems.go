// SPDX-License-Identifier: MIT

// Package problems 提供对 Problem 相关内容的管理
package problems

import (
	"fmt"

	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"
)

type Problems struct {
	prefix   string
	problems []*StatusProblem // 不用 map，保证 Problems 顺序相同。
}

type StatusProblem struct {
	id string // 用户指定的原始值

	Type   string // 带 prefix
	Status int
	Title  localeutil.LocaleStringer
	Detail localeutil.LocaleStringer
}

func New(prefix string) *Problems {
	p := &Problems{
		prefix:   prefix,
		problems: make([]*StatusProblem, 0, 100),
	}

	for id, status := range statuses {
		text := "problem." + id
		p.Add(id, status, localeutil.Phrase(text), localeutil.Phrase(text+".detail"))
	}

	return p
}

func (p *Problems) Add(id string, status int, title, detail localeutil.LocaleStringer) {
	if p.exists(id) {
		panic(fmt.Sprintf("存在相同值的 id 参数 %s", id))
	}
	if !IsValidStatus(status) {
		panic("status 必须是一个有效的状态码")
	}

	if title == nil {
		panic("title 不能为空")
	}

	s := &StatusProblem{id: id, Status: status, Title: title, Detail: detail}
	if p.prefix == ProblemAboutBlank {
		s.Type = ProblemAboutBlank
	} else {
		s.Type = p.prefix + s.id
	}
	p.problems = append(p.problems, s)
}

func (p *Problems) exists(id string) bool {
	return sliceutil.Exists(p.problems, func(sp *StatusProblem, _ int) bool { return sp.id == id })
}

func (p *Problems) Visit(visit func(prefix, id string, status int, title, detail localeutil.LocaleStringer)) {
	for _, s := range p.problems {
		visit(p.prefix, s.id, s.Status, s.Title, s.Detail)
	}
}

func (p *Problems) Problem(id string) *StatusProblem {
	sp, found := sliceutil.At(p.problems, func(sp *StatusProblem, _ int) bool { return sp.id == id })
	if !found { // 初始化时没有给定相关的定义，所以直接 panic。
		panic(fmt.Sprintf("未找到有关 %s 的定义", id))
	}
	return sp
}

func IsValidStatus(status int) bool { return status >= 100 && status < 600 }
