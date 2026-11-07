// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package problems

import (
	"slices"
	"testing"

	"github.com/issue9/assert/v6"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type problem struct {
	Type  string
	Title string
}

func builder(lp *Item, printer *localeutil.Printer, p *problem) {
	p.Type = lp.Type
	p.Title = lp.Title.LocaleString(printer)
}

func TestProblems_Add(t *testing.T) {
	a := assert.New(t, false)

	ps := New(builder)
	a.NotNil(ps)
	l := len(ps.problems)

	a.False(ps.Exists("40010")).
		False(ps.Exists("40011")).
		True(ps.Exists(NotFound))

	ps.Add("", []*Item{
		{Status: 400, ID: "40010", Title: localeutil.Phrase("title"), Detail: localeutil.Phrase("detail")},
		{Status: 400, ID: "40011", Title: localeutil.Phrase("title"), Detail: localeutil.Phrase("detail")},
	}...)
	a.True(ps.Exists("40010")).
		True(ps.Exists("40011")).
		Equal(l+2, len(ps.problems)).
		Length(slices.Collect(ps.All()), l+2)

	a.PanicString(func() {
		ps.Add("", []*Item{
			{Status: 400, ID: "40010", Title: localeutil.Phrase("title"), Detail: localeutil.Phrase("detail")},
		}...)
	}, "该 ID 值 40010 已经存在")

	a.PanicString(func() {
		ps.Add("", []*Item{
			{Status: 99, ID: "40012", Title: localeutil.Phrase("title"), Detail: localeutil.Phrase("detail")},
		}...)
	}, "status 必须是一个有效的状态码")

	a.PanicString(func() {
		ps.Add("", []*Item{
			{Status: 400, ID: "40013", Title: nil, Detail: nil},
		}...)
	}, "title 不能为空")
}

func TestProblems_Init(t *testing.T) {
	a := assert.New(t, false)
	p := message.NewPrinter(language.SimplifiedChinese)

	ps := New(builder)
	a.NotNil(ps)
	ps.Add("", &Item{Status: 400, ID: "40010", Title: localeutil.Phrase("title"), Detail: localeutil.Phrase("detail")})
	pp := &problem{}
	ps.Init("40010", p, pp)
	a.Equal(pp.Type, AboutBlank)

	ps = New(builder)
	a.NotNil(ps)
	ps.Add("", &Item{Status: 400, ID: "40011", Type: "https://example.com/qa#40011", Title: localeutil.Phrase("title"), Detail: localeutil.Phrase("detail")})
	ps.Add("https://example.com/qa#", &Item{Status: 400, ID: "40012", Type: "40012", Title: localeutil.Phrase("title"), Detail: localeutil.Phrase("detail")})
	pp = &problem{}
	ps.Init("40011", p, pp)
	a.Equal(pp.Type, "https://example.com/qa#40011")
	pp = &problem{}
	ps.Init("40012", p, pp)
	a.Equal(pp.Type, "https://example.com/qa#40012")

	a.PanicString(func() {
		ps.Init("not-exists", p, pp)
	}, "未找到有关 not-exists 的定义")
}
