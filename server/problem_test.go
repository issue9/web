// SPDX-License-Identifier: MIT

package server

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"

	"github.com/issue9/web/problem"
)

var _ Responser = &Problem{}

func TestProblems_Add(t *testing.T) {
	a := assert.New(t, false)

	ps := newProblems()
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

	ps := newProblems()
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

func TestProblem_Apply(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, nil)
	s.Problems().Add("40010", 400, localeutil.Phrase("40010"), localeutil.Phrase("lang"))

	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Header("accept", "application/json;charset=utf-8").Request()
	ctx := s.newContext(w, r, nil)
	p := ctx.Problem("40010", nil)
	p.Apply(ctx)
	a.Equal(w.Result().StatusCode, 400)
	pp := &problem.RFC7807{}
	a.NotError(json.Unmarshal(w.Body.Bytes(), pp))
	a.Equal(pp.Detail, "und").
		Equal(pp.Type, "40010").
		Equal(pp.Title, "40010").
		Equal(pp.Status, 400).
		Empty(pp.Instance)

	s.Problems().SetTypeBaseURL("https://example.com/problems/")
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept", "application/json;charset=utf-8").
		Header("accept-language", language.SimplifiedChinese.String()).
		Request()
	ctx = s.newContext(w, r, nil)
	p = ctx.Problem("40010", nil).WithTitle("title").WithInstance("/instance")
	p.Apply(ctx)
	a.Equal(w.Result().StatusCode, 400)
	pp = &problem.RFC7807{}
	a.NotError(json.Unmarshal(w.Body.Bytes(), pp))
	a.Equal(pp.Detail, "hans").
		Equal(pp.Type, "https://example.com/problems/40010").
		Equal(pp.Title, "title").
		Equal(pp.Status, 400).
		Equal(pp.Instance, "/instance")
}
