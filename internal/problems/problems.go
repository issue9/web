// SPDX-License-Identifier: MIT

// Package problems 提供对 Problem 管理
package problems

import (
	"fmt"
	"net/http"

	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"
	"golang.org/x/text/message"
)

const AboutBlank = "about:blank"

// Problems 管理 Problem
//
// P 表示 Problem 接口类型
type Problems[P any] struct {
	builder    func(id, title string, status int) P
	typePrefix string
	problems   []*StatusProblem // 不用 map，保证 Problems 顺序相同。
}

type StatusProblem struct {
	t string // 实际的 type 值，id 仅用于查找

	ID     string
	Status int
	Title  localeutil.LocaleStringer
	Detail localeutil.LocaleStringer
}

func New[P any](f func(id, title string, status int) P) *Problems[P] {
	p := &Problems[P]{
		builder:  f,
		problems: make([]*StatusProblem, 0, 50),
	}

	for id, status := range statuses {
		msg := localeutil.Phrase(http.StatusText(status))
		p.Add(&StatusProblem{ID: id, Status: status, Title: msg, Detail: msg})
	}

	return p
}

func (p *Problems[P]) TypePrefix() string { return p.typePrefix }

func (p *Problems[P]) SetTypePrefix(prefix string) {
	p.typePrefix = prefix
	if prefix == AboutBlank {
		for _, s := range p.problems {
			s.t = ""
		}
		return
	}

	for _, s := range p.problems {
		s.t = prefix + s.ID
	}
}

func (p *Problems[P]) Add(s ...*StatusProblem) {
	for _, sp := range s {
		p.add(sp)
	}
}

func (p *Problems[P]) add(s *StatusProblem) {
	if p.Exists(s.ID) {
		panic(fmt.Sprintf("存在相同值的 id 参数 %s", s.ID))
	}

	if p.typePrefix == AboutBlank {
		s.t = ""
	} else {
		s.t = p.typePrefix + s.ID
	}
	p.problems = append(p.problems, s)
}

func (p *Problems[P]) Exists(id string) bool {
	return sliceutil.Exists(p.problems, func(sp *StatusProblem) bool { return sp.ID == id })
}

func (p *Problems[P]) Problems() []*StatusProblem { return p.problems }

func (p *Problems[P]) Problem(printer *message.Printer, id string) P {
	sp, found := sliceutil.At(p.problems, func(sp *StatusProblem) bool { return sp.ID == id })
	if !found { // 初始化时没有给定相关的定义，所以直接 panic。
		panic(fmt.Sprintf("未找到有关 %s 的定义", id))
	}

	return p.builder(sp.t, sp.Title.LocaleString(printer), sp.Status)
}
