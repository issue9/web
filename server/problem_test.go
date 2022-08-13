// SPDX-License-Identifier: MIT

package server

import (
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/assert/v3/rest"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
)

func TestProblems_Add(t *testing.T) {
	a := assert.New(t, false)

	ps := newProblems(RFC7807Builder)
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

	ps := newProblems(RFC7807Builder)
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

func TestProblems_Mimetype(t *testing.T) {
	a := assert.New(t, false)
	ps := newProblems(RFC7807Builder)
	a.NotNil(ps)

	a.Equal(ps.mimetype("application/json"), "application/json")
	ps.AddMimetype("application/json", "application/problem+json")
	a.Equal(ps.mimetype("application/json"), "application/problem+json")
	a.PanicString(func() {
		ps.AddMimetype("application/json", "application/problem")
	}, "已经存在的 mimetype")
}

func TestProblems_Problem(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, nil)
	ps := s.Problems()

	a.PanicString(func() {
		ps.Problem("not-exists")
	}, "未找到有关 not-exists 的定义")

	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").
		Header("Accept", "application/json").
		Header("accept-language", language.SimplifiedChinese.String()).
		Request()
	ctx := s.newContext(w, r, nil)

	p := ps.Problem("41110")
	a.NotNil(p)
	p.Apply(ctx)
	a.Equal(w.Body.String(), `{"type":"41110","title":"hans","status":411}`).
		Equal(w.Result().StatusCode, 411)
}
