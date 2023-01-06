// SPDX-License-Identifier: MIT

package problems

import (
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func builder(id, title string, status int) string {
	return id
}

func TestProblems_Add_Problems(t *testing.T) {
	a := assert.New(t, false)

	ps := New(builder)
	a.NotNil(ps)
	l := len(ps.Problems())

	a.False(ps.Exists("40010")).
		False(ps.Exists("40011"))
	ps.Add(
		&StatusProblem{ID: "40010", Status: 400, Title: localeutil.Phrase("title"), Detail: localeutil.Phrase("detail")},
		&StatusProblem{ID: "40011", Status: 400, Title: localeutil.Phrase("title"), Detail: localeutil.Phrase("detail")},
	)
	a.True(ps.Exists("40010")).
		True(ps.Exists("40011")).
		Equal(l+2, len(ps.Problems()))

	a.PanicString(func() {
		ps.Add(&StatusProblem{ID: "40010", Status: 400, Title: localeutil.Phrase("title"), Detail: localeutil.Phrase("detail")})
	}, "存在相同值的 id 参数")

	a.PanicString(func() {
		ps.Add(&StatusProblem{ID: "40012", Status: 99, Title: localeutil.Phrase("title"), Detail: localeutil.Phrase("detail")})
	}, "status 必须是一个有效的状态码")

	a.PanicString(func() {
		ps.Add(&StatusProblem{ID: "40012", Status: 200, Title: nil, Detail: nil})
	}, "title 不能为空")
}

func TestProblems_TypePrefix_Problem(t *testing.T) {
	a := assert.New(t, false)
	ps := New(builder)
	a.NotNil(ps)
	printer := message.NewPrinter(language.Chinese)

	ps.Add(
		&StatusProblem{ID: "40010", Status: 400, Title: localeutil.Phrase("title"), Detail: localeutil.Phrase("detail")},
	)
	a.Equal(ps.Problem(printer, "40010"), "40010").
		Equal(ps.TypePrefix(), "")

	ps.SetTypePrefix("https://example.com/qa#")
	ps.Add(
		&StatusProblem{ID: "40011", Status: 400, Title: localeutil.Phrase("title"), Detail: localeutil.Phrase("detail")},
	)
	a.Equal(ps.Problem(printer, "40010"), "https://example.com/qa#40010").
		Equal(ps.Problem(printer, "40011"), "https://example.com/qa#40011").
		Equal(ps.TypePrefix(), "https://example.com/qa#")

	ps.SetTypePrefix(ProblemAboutBlank)
	ps.Add(
		&StatusProblem{ID: "40012", Status: 400, Title: localeutil.Phrase("title"), Detail: localeutil.Phrase("detail")},
	)
	a.Equal(ps.Problem(printer, "40010"), ProblemAboutBlank).
		Equal(ps.Problem(printer, "40012"), ProblemAboutBlank).
		Equal(ps.TypePrefix(), ProblemAboutBlank)

	a.PanicString(func() {
		ps.Problem(printer, "not-exists")
	}, "未找到有关 not-exists 的定义")
}

func TestProblems_Status(t *testing.T) {
	a := assert.New(t, false)
	ps := New(builder)
	a.NotNil(ps)

	s := ps.Status(201)
	s.Add("40010", localeutil.Phrase("title"), localeutil.Phrase("detail")).
		Add("40011", localeutil.Phrase("title"), localeutil.Phrase("detail"))
	item, found := sliceutil.At(ps.problems, func(p *StatusProblem) bool { return p.ID == "40010" })
	a.True(found).Equal(item.Status, 201)

	a.PanicString(func() {
		ps.Status(99)
	}, "无效的状态码 99")
}
