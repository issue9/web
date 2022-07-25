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

func TestProblems_Problem(t *testing.T) {
	a := assert.New(t, false)
	ps := newProblems()
	ps.Add("40010", 400, localeutil.Phrase("40010"), localeutil.Phrase("40010"))

	a.PanicString(func() {
		ps.Problem("not-exists", nil)
	}, "未找到有关 not-exists 的定义")

	p := ps.Problem("40010", nil)
	a.NotNil(p).
		Equal(p.id, "40010").
		NotNil(p.p).
		TypeEqual(true, p.p, problem.NewRFC7807Problem()).
		Zero(p.p.GetStatus()).
		Empty(p.p.GetType()).
		Empty(p.p.GetTitle()).
		Empty(p.p.GetDetail()).
		Empty(p.p.GetInstance())

	p.WithTitle("title").WithStatus(201).WithDetail("detail").WithInstance("/instance")
	a.NotNil(p).
		Equal(p.id, "40010").
		Empty(p.p.GetType()).
		Equal(p.p.GetStatus(), 201).
		Equal(p.p.GetTitle(), "title").
		Equal(p.p.GetDetail(), "detail").
		Equal(p.p.GetInstance(), "/instance")

	p = ps.Problem("40010", &problem.InvalidParamsProblem{})
	a.NotNil(p).
		Equal(p.id, "40010").
		NotNil(p.p).
		TypeEqual(true, p.p, &problem.InvalidParamsProblem{}).
		Zero(p.p.GetStatus()).
		Empty(p.p.GetType()).
		Empty(p.p.GetTitle()).
		Empty(p.p.GetDetail()).
		Empty(p.p.GetInstance())
}

func TestProblem_Apply(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, nil)
	s.Problems().Add("40010", 400, localeutil.Phrase("40010"), localeutil.Phrase("lang"))

	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Header("accept", "application/json;charset=utf-8").Request()
	p := s.Problems().Problem("40010", nil)
	ctx := s.newContext(w, r, nil)
	p.Apply(ctx)
	a.Equal(w.Result().StatusCode, 400)
	pp := problem.NewRFC7807Problem()
	a.NotError(json.Unmarshal(w.Body.Bytes(), pp))
	a.Equal(pp.GetDetail(), "und").
		Equal(pp.GetType(), "40010").
		Equal(pp.GetTitle(), "40010").
		Equal(pp.GetStatus(), 400).
		Empty(pp.GetInstance())

	s.Problems().SetInstanceBaseURL("https://example.com/instance/")
	s.Problems().SetTypeBaseURL("https://example.com/problems/")
	p = s.Problems().Problem("40010", nil).WithTitle("title").WithInstance("instance")
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").
		Header("accept", "application/json;charset=utf-8").
		Header("accept-language", language.SimplifiedChinese.String()).
		Request()
	ctx = s.newContext(w, r, nil)
	p.Apply(ctx)
	a.Equal(w.Result().StatusCode, 400)
	pp = problem.NewRFC7807Problem()
	a.NotError(json.Unmarshal(w.Body.Bytes(), pp))
	a.Equal(pp.GetDetail(), "hans").
		Equal(pp.GetType(), "https://example.com/problems/40010").
		Equal(pp.GetTitle(), "title").
		Equal(pp.GetStatus(), 400).
		Equal(pp.GetInstance(), "https://example.com/instance/instance")
}
