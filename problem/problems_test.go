// SPDX-License-Identifier: MIT

package problem

import (
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

func TestProblems_Add(t *testing.T) {
	a := assert.New(t, false)

	ps := NewProblems(RFC7807Builder, "", false)
	a.NotNil(ps)
	a.Equal(0, len(ps.problems))

	ps.Add("40010", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	a.Equal(1, len(ps.problems))

	ps.Add("40011", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	a.Equal(2, len(ps.problems))

	a.PanicString(func() {
		ps.Add("40010", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	}, "存在相同值的 id 参数")
	a.Equal(2, len(ps.problems))
}

func TestProblems_Visit(t *testing.T) {
	a := assert.New(t, false)

	ps := NewProblems(RFC7807Builder, "", false)
	cnt := 0
	ps.Visit(func(id string, status int, title, detail localeutil.LocaleStringer) bool {
		cnt++
		return true
	})
	a.Equal(0, cnt)

	ps.Add("40010", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	cnt = 0
	ps.Visit(func(id string, status int, title, detail localeutil.LocaleStringer) bool {
		cnt++
		return true
	})
	a.Equal(1, cnt)

	ps.Add("40011", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	ps.Add("40012", 400, localeutil.Phrase("title"), localeutil.Phrase("detail"))
	cnt = 0
	ps.Visit(func(id string, status int, title, detail localeutil.LocaleStringer) bool {
		cnt++
		return true
	})
	a.Equal(3, cnt)

	cnt = 0
	ps.Visit(func(id string, status int, title, detail localeutil.LocaleStringer) bool {
		cnt++
		return false // 中断
	})
	a.Equal(1, cnt)
}

func TestProblems_Problem(t *testing.T) {
	a := assert.New(t, false)

	cat := catalog.NewBuilder()
	cat.SetString(language.SimplifiedChinese, "lang", "hans")
	cnp := message.NewPrinter(language.SimplifiedChinese, message.Catalog(cat))
	twp := message.NewPrinter(language.TraditionalChinese, message.Catalog(cat))

	ps := NewProblems(RFC7807Builder, "", false)
	ps.Add("40010", 400, localeutil.Phrase("lang"), localeutil.Phrase("40010"))
	ps.Add("40011", 400, localeutil.Phrase("lang"), localeutil.Phrase("40011"))

	a.PanicString(func() {
		ps.Problem("not-exists", nil)
	}, "未找到有关 not-exists 的定义")

	p := ps.Problem("40010", cnp)
	a.NotNil(p).Equal(p.Status(), 400)
	pp, ok := p.(*rfc7807)
	a.True(ok).NotNil(pp).Equal(pp.title, "hans")

	p = ps.Problem("40010", twp)
	a.NotNil(p).Equal(p.Status(), 400)
	pp, ok = p.(*rfc7807)
	a.True(ok).NotNil(pp).Equal(pp.title, "lang")
}
