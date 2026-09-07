// SPDX-FileCopyrightText: 2024-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/issue9/assert/v5"
	"github.com/issue9/localeutil"
	"github.com/issue9/mux/v9/types"
)

func buildMinValidator(v int) func(int) bool { return func(a int) bool { return a >= v } }

func buildMaxValidator(v int) func(int) bool { return func(a int) bool { return a < v } }

func trimRight(v *string) { *v = strings.TrimRight(*v, " ") }

func upper(v *string) { *v = strings.ToUpper(*v) }

func zero[T any](v T) bool { return reflect.ValueOf(v).IsZero() }

func required[T any](v T) bool { return !zero(v) }

func newFilter(a *assert.Assertion) *FilterContext {
	s := newTestServer(a)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	ctx := s.NewContext(w, r, types.NewContext())
	return ctx.NewFilterContext(false)
}

func TestSanitizerRule(t *testing.T) {
	a := assert.New(t, false)

	id := "x "
	v := newFilter(a).Add("id", &id, SanitizerRule(trimRight), SanitizerRule(upper))
	a.Length(v.problem.Params, 0).Equal(id, "X")
}

func TestValidatorRule(t *testing.T) {
	a := assert.New(t, false)

	id := " "
	v := newFilter(a).Add("id", &id, ValidatorRule[string](required, localeutil.Phrase("required")))
	a.Nil(v.problem.Params).Equal(id, " ")

	id = "X "
	v = newFilter(a).Add("id", &id, SanitizerRule(upper), ValidatorRule[string](required, localeutil.Phrase("required")))
	a.Nil(v.problem.Params).Equal(id, "X ")
}

func TestSliceSanitizerRule(t *testing.T) {
	a := assert.New(t, false)

	vals := []string{"s1 ", "s2"}
	v := newFilter(a).Add("vals", &vals, SliceSanitizerRule[[]string](trimRight, upper))
	a.Nil(v.problem.Params).Equal(vals, []string{"S1", "S2"})
}

func TestMapSanitizerRule(t *testing.T) {
	a := assert.New(t, false)

	vals := map[string]string{"s1 ": "s1 ", "s2": "s2"}
	v := newFilter(a).Add("vals", &vals, MapSanitizerRule[map[string]string](upper))
	a.Nil(v.problem.Params).Equal(vals, map[string]string{"s1 ": "S1 ", "s2": "S2"})
}

func TestSliceValidatorRule(t *testing.T) {
	a := assert.New(t, false)

	vals := []string{"s1 ", "s2"}
	v := newFilter(a).Add("vals", &vals, SliceValidatorRule[[]string](required, localeutil.Phrase("required")))
	a.Nil(v.problem.Params).Equal(vals, []string{"s1 ", "s2"})

	vals = []string{"s1 ", ""}
	v = newFilter(a).Add("vals", &vals, SliceValidatorRule[[]string](required, localeutil.Phrase("required")))
	a.Equal(v.problem.Params[0].Reason, localeutil.Phrase("required")).
		Equal(v.problem.Params[0].Name, "vals[1]").
		Equal(vals, []string{"s1 ", ""})
}

func TestMapValidatorRule(t *testing.T) {
	a := assert.New(t, false)

	vals := map[string]string{"s1 ": "s1", "s2": "s2"}
	v := newFilter(a).Add("vals", &vals, MapValidatorRule[map[string]string](required, localeutil.Phrase("required")))
	a.Nil(v.problem.Params)

	vals = map[string]string{"s1 ": "x", "s2": ""}
	v = newFilter(a).Add("vals", &vals, MapValidatorRule[map[string]string](required, localeutil.Phrase("required")))
	a.Equal(v.problem.Params[0].Reason, localeutil.Phrase("required")).
		Equal(v.problem.Params[0].Name, "vals[s2]").
		Equal(vals, map[string]string{"s1 ": "x", "s2": ""})
}
