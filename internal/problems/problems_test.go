// SPDX-License-Identifier: MIT

package problems

import (
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func builder(id string, status int, title, detail string) string { return id }

func TestProblems_Add_Problems(t *testing.T) {
	a := assert.New(t, false)

	ps := New("", builder)
	a.NotNil(ps)
	l := len(ps.problems)
	a.Equal(l, len(statuses))

	a.False(ps.exists("40010")).
		False(ps.exists("40011"))
	ps.Add("40010", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	ps.Add("40011", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	a.True(ps.exists("40010")).
		True(ps.exists("40011")).
		Equal(l+2, len(ps.problems))

	a.PanicString(func() {
		ps.Add("40010", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	}, "存在相同值的 id 参数")

	a.PanicString(func() {
		ps.Add("40012", 99, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	}, "status 必须是一个有效的状态码")

	a.PanicString(func() {
		ps.Add("40012", 200, nil, nil)
	}, "title 不能为空")
}

func TestProblems_Problem(t *testing.T) {
	a := assert.New(t, false)
	printer := message.NewPrinter(language.Chinese)

	ps := New("", builder)
	a.NotNil(ps)
	ps.Add("40010", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	a.Equal(ps.Problem(printer, "40010"), "40010")

	ps = New("https://example.com/qa#", builder)
	a.NotNil(ps)
	ps.Add("40011", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	a.Equal(ps.Problem(printer, "40011"), "https://example.com/qa#40011")

	ps = New(ProblemAboutBlank, builder)
	a.NotNil(ps)
	ps.Add("40012", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	a.Equal(ps.Problem(printer, "40012"), ProblemAboutBlank)

	a.PanicString(func() {
		ps.Problem(printer, "not-exists")
	}, "未找到有关 not-exists 的定义")
}
