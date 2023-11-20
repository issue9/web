// SPDX-License-Identifier: MIT

package server

import (
	"fmt"

	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/status"
)

type problems struct {
	prefix   string
	problems []*problem // 不能用 map，需保证元素的顺序相同。
}

type problem struct {
	id string // 用户指定的原始值

	Type   string // 带 prefix
	Status int
	Title  web.LocaleStringer
	Detail web.LocaleStringer
}

func newProblems(prefix string) *problems {
	ps := &problems{
		prefix:   prefix,
		problems: make([]*problem, 0, 100),
	}
	initProblems(ps)

	return ps
}

func (srv *httpServer) Problems() web.Problems { return srv.problems }

func (ps *problems) Prefix() string { return ps.prefix }

func (ps *problems) Add(s int, p ...web.LocaleProblem) web.Problems {
	if !status.IsProblemStatus(s) { // 只需验证大于 400 的状态码。
		panic("status 必须是一个有效的状态码")
	}

	for _, pp := range p {
		ps.add(pp.ID, s, pp.Title, pp.Detail)
	}

	return ps
}

func (ps *problems) add(id string, s int, title, detail web.LocaleStringer) {
	if ps.exists(id) {
		panic(fmt.Sprintf("存在相同值的 id 参数 %s", id))
	}

	if title == nil {
		panic("title 不能为空")
	}

	p := &problem{id: id, Status: s, Title: title, Detail: detail}
	if ps.prefix == web.ProblemAboutBlank {
		p.Type = web.ProblemAboutBlank
	} else {
		p.Type = ps.prefix + p.id
	}
	ps.problems = append(ps.problems, p)
}

func (ps *problems) exists(id string) bool {
	return sliceutil.Exists(ps.problems, func(p *problem, _ int) bool { return p.id == id })
}

func (ps *problems) Visit(visit func(id string, status int, title, detail web.LocaleStringer)) {
	for _, s := range ps.problems {
		visit(s.id, s.Status, s.Title, s.Detail)
	}
}

func (ps *problems) Init(pp *web.RFC7807, id string, p *localeutil.Printer) {
	sp, found := sliceutil.At(ps.problems, func(p *problem, _ int) bool { return p.id == id })
	if !found { // 初始化时没有给定相关的定义，所以直接 panic。
		panic(fmt.Sprintf("未找到有关 %s 的定义", id))
	}

	pp.Init(sp.Type, sp.Title.LocaleString(p), sp.Detail.LocaleString(p), sp.Status)
}
