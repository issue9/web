// SPDX-License-Identifier: MIT

package server

import (
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/assert/v3/rest"
	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"

	"github.com/issue9/web/filter"
)

type object struct {
	Name string
	Age  int
}

func required[T any](v T) bool { return !reflect.ValueOf(v).IsZero() }

func min(v int) filter.ValidatorFuncOf[int] {
	return func(a int) bool { return a >= v }
}

func max(v int) filter.ValidatorFuncOf[int] {
	return func(a int) bool { return a < v }
}

func in(element ...int) filter.ValidatorFuncOf[int] {
	return func(v int) bool {
		return sliceutil.Exists(element, func(elem int) bool { return elem == v })
	}
}

func notIn(element ...int) filter.ValidatorFuncOf[int] {
	return func(v int) bool {
		return !sliceutil.Exists(element, func(elem int) bool { return elem == v })
	}
}

func TestContext_NewFilter(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a, nil)
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	ctx := s.newContext(w, r, nil)

	min_2 := filter.NewRule(min(-2), localeutil.Phrase("-2"))
	min_3 := filter.NewRule(min(-3), localeutil.Phrase("-3"))
	max50 := filter.NewRule(max(50), localeutil.Phrase("50"))
	max_4 := filter.NewRule(max(-4), localeutil.Phrase("-4"))

	n100 := -100
	p100 := 100
	v := ctx.NewFilter(false).
		AddFilter(filter.New(filter.NewRules(min_2, min_3))("f1", &n100)).
		AddFilter(filter.New(filter.NewRules(max50, max_4))("f2", &p100))
	a.Equal(v.keys, []string{"f1", "f2"}).
		Equal(v.reasons, []string{"-2", "50"})

	n100 = -100
	p100 = 100
	v = ctx.NewFilter(true).
		AddFilter(filter.New(filter.NewRules(min_2, min_3))("f1", &n100)).
		AddFilter(filter.New(filter.NewRules(max50, max_4))("f2", &p100))
	a.Equal(v.keys, []string{"f1"}).
		Equal(v.reasons, []string{"-2"})
}

func TestFilter_When(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a, nil)
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	ctx := s.newContext(w, r, nil)

	min18 := filter.NewRule(min(18), localeutil.Phrase("不能小于 18"))
	req := filter.ValidatorFuncOf[string](required[string])
	notEmpty := filter.NewRule(req, localeutil.Phrase("不能为空"))

	obj := &object{}
	v := ctx.NewFilter(false).
		AddFilter(filter.New(min18)("obj/age", &obj.Age)).
		When(obj.Age > 0, func(v *Filter) {
			v.AddFilter(filter.New(notEmpty)("obj/name", &obj.Name))
		})
	a.Equal(v.keys, []string{"obj/age"}).
		Equal(v.reasons, []string{"不能小于 18"})

	obj = &object{Age: 15}
	v = ctx.NewFilter(false).
		AddFilter(filter.New(min18)("obj/age", &obj.Age)).
		When(obj.Age > 0, func(v *Filter) {
			v.AddFilter(filter.New(notEmpty)("obj/name", &obj.Name))
		})
	a.Equal(v.keys, []string{"obj/age", "obj/name"}).
		Equal(v.reasons, []string{"不能小于 18", "不能为空"})
}
