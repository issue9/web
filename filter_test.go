// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v5"
	"github.com/issue9/mux/v9/types"
)

func TestFilterContext(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	ctx := s.NewContext(w, r, types.NewContext())

	min2 := ValidatorRule(buildMinValidator(-2), Phrase("-2"))
	min3 := ValidatorRule(buildMinValidator(-3), Phrase("-3"))
	max50 := ValidatorRule(buildMaxValidator(50), Phrase("50"))
	max4 := ValidatorRule(buildMaxValidator(-4), Phrase("-4"))

	n100 := -100
	p100 := 100
	v := ctx.NewFilterContext(false).
		Add("f1", &n100, min2, min3).
		Add("f2", &p100, max50, max4)
	a.Equal(v.problem.Params, []ProblemParam{
		{Name: "f1", Reason: "-2"},
		{Name: "f2", Reason: "50"},
	})

	n100 = -100
	p100 = 100
	v = ctx.NewFilterContext(true).
		Add("f1", &n100, min2, min3).
		Add("f2", &p100, max50, max4)
	a.Equal(v.problem.Params, []ProblemParam{
		{Name: "f1", Reason: "-2"},
	})
}

func TestFilterContext_New(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	ctx := s.NewContext(w, r, types.NewContext())

	v := ctx.NewFilterContext(false)
	v1 := v.New("v1.", func(f *FilterContext) {
		f.AddReason("f1", StringPhrase("s1"))
		v2 := f.New("v2.", func(f *FilterContext) {
			f.AddError("f2", errors.New("s2"))
		})
		a.Equal(v2, f)
	})
	a.Equal(v1, v)

	a.Equal(v.problem.Params, []ProblemParam{
		{Name: "v1.f1", Reason: "s1"},
		{Name: "v1.v2.f2", Reason: "s2"},
	})
}

func TestFilterContext_When(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	ctx := s.NewContext(w, r, types.NewContext())

	min18 := ValidatorRule(buildMinValidator(18), Phrase("不能小于 18"))
	notEmpty := ValidatorRule(required[string], Phrase("不能为空"))

	obj := &object{}
	v := ctx.NewFilterContext(false).
		Add("obj/age", &obj.Age, min18).
		When(obj.Age > 0, func(v *FilterContext) {
			v.Add("obj/name", &obj.Name, notEmpty)
		})
	a.Equal(v.problem.Params, []ProblemParam{
		{Name: "obj/age", Reason: "不能小于 18"},
	})

	obj = &object{Age: 15}
	v = ctx.NewFilterContext(false).
		Add("obj/age", &obj.Age, min18).
		When(obj.Age > 0, func(v *FilterContext) {
			v.Add("obj/name", &obj.Name, notEmpty)
		})
	a.Equal(v.problem.Params, []ProblemParam{
		{Name: "obj/age", Reason: "不能小于 18"},
		{Name: "obj/name", Reason: "不能为空"},
	})
}
