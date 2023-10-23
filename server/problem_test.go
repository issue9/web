// SPDX-License-Identifier: MIT

package server

import (
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/web"
)

var _ web.Problems = &problems{}

func TestProblems_Add(t *testing.T) {
	a := assert.New(t, false)

	ps := newProblems("")
	a.NotNil(ps)
	l := len(ps.problems)

	a.False(ps.exists("40010")).
		False(ps.exists("40011")).
		True(ps.exists(web.ProblemNotFound))

	ps.Add(400, []web.LocaleProblem{
		{ID: "40010", Title: localeutil.Phrase("title"), Detail: localeutil.Phrase("detail")},
		{ID: "40011", Title: localeutil.Phrase("title"), Detail: localeutil.Phrase("detail")},
	}...)
	a.True(ps.exists("40010")).
		True(ps.exists("40011")).
		Equal(l+2, len(ps.problems))

	a.PanicString(func() {
		ps.Add(400, []web.LocaleProblem{
			{ID: "40010", Title: localeutil.Phrase("title"), Detail: localeutil.Phrase("detail")},
		}...)
	}, "存在相同值的 id 参数")

	a.PanicString(func() {
		ps.Add(99, []web.LocaleProblem{
			{ID: "40012", Title: localeutil.Phrase("title"), Detail: localeutil.Phrase("detail")},
		}...)
	}, "status 必须是一个有效的状态码")

	a.PanicString(func() {
		ps.Add(412, []web.LocaleProblem{
			{ID: "40013", Title: nil, Detail: nil},
		}...)
	}, "title 不能为空")
}

func TestProblems_Init(t *testing.T) {
	a := assert.New(t, false)
	p := message.NewPrinter(language.SimplifiedChinese)

	ps := newProblems("")
	a.NotNil(ps)
	ps.Add(400, web.LocaleProblem{ID: "40010", Title: localeutil.Phrase("title"), Detail: localeutil.Phrase("detail")})
	pp := &web.RFC7807{}
	ps.Init(pp, "40010", p)
	a.Equal(pp.Type, "40010")

	ps = newProblems("https://example.com/qa#")
	a.NotNil(ps)
	ps.Add(400, web.LocaleProblem{ID: "40011", Title: localeutil.Phrase("title"), Detail: localeutil.Phrase("detail")})
	pp = &web.RFC7807{}
	ps.Init(pp, "40011", p)
	a.Equal(pp.Type, "https://example.com/qa#40011").
		Equal(ps.Prefix(), "https://example.com/qa#")

	ps = newProblems(web.ProblemAboutBlank)
	a.NotNil(ps)
	ps.Add(400, web.LocaleProblem{ID: "40012", Title: localeutil.Phrase("title"), Detail: localeutil.Phrase("detail")})
	pp = &web.RFC7807{}
	ps.Init(pp, "40012", p)
	a.Equal(pp.Type, web.ProblemAboutBlank)

	a.PanicString(func() {
		ps.Init(pp, "not-exists", p)
	}, "未找到有关 not-exists 的定义")
}
