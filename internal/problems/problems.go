// SPDX-License-Identifier: MIT

//go:generate go run ./make_id.go

// Package problems 提供对 Problem 相关内容的管理
package problems

import (
	"fmt"

	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"
)

type Problems struct {
	prefix   string
	problems []*Problem // 不用 map，保证 Problems 顺序相同。
}

type Problem struct {
	id string // 用户指定的原始值

	Type   string // 带 prefix
	Status int
	Title  localeutil.LocaleStringer
	Detail localeutil.LocaleStringer
}

func New(prefix string) *Problems {
	p := &Problems{
		prefix:   prefix,
		problems: make([]*Problem, 0, 100),
	}
	p.initLocales()
	return p
}

func (ps *Problems) Add(id string, status int, title, detail localeutil.LocaleStringer) {
	if ps.exists(id) {
		panic(fmt.Sprintf("存在相同值的 id 参数 %s", id))
	}
	if !validProblemStatus(status) {
		panic("status 必须是一个有效的状态码")
	}

	if title == nil {
		panic("title 不能为空")
	}

	p := &Problem{id: id, Status: status, Title: title, Detail: detail}
	if ps.prefix == ProblemAboutBlank {
		p.Type = ProblemAboutBlank
	} else {
		p.Type = ps.prefix + p.id
	}
	ps.problems = append(ps.problems, p)
}

func (ps *Problems) exists(id string) bool {
	return sliceutil.Exists(ps.problems, func(p *Problem, _ int) bool { return p.id == id })
}

func (ps *Problems) Visit(visit func(prefix, id string, status int, title, detail localeutil.LocaleStringer)) {
	for _, s := range ps.problems {
		visit(ps.prefix, s.id, s.Status, s.Title, s.Detail)
	}
}

func (ps *Problems) Problem(id string) *Problem {
	sp, found := sliceutil.At(ps.problems, func(p *Problem, _ int) bool { return p.id == id })
	if !found { // 初始化时没有给定相关的定义，所以直接 panic。
		panic(fmt.Sprintf("未找到有关 %s 的定义", id))
	}
	return sp
}

// [Problems] 中也没有处理小于 400 的状态码，所以此处验证也不理会小于 400 的值。
func validProblemStatus(status int) bool { return ids[status] != "" }

func Status(id string) int { return statuses[id] }

func ID(status int) string { return ids[status] }
