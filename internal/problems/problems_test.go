// SPDX-License-Identifier: MIT

package problems

import (
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
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
