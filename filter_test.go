// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/issue9/assert/v3"
)

type obj2 struct {
	o1 *object
	o2 object
}

func trimRight(v *string) { *v = strings.TrimRight(*v, " ") }

func zero[T any](v T) bool { return reflect.ValueOf(v).IsZero() }

func required[T any](v T) bool { return !zero(v) }

func between[T ~int | ~uint | float32 | float64](min, max T) func(T) bool {
	return func(vv T) bool { return vv >= min && vv <= max }
}

func min(v int) func(int) bool { return func(a int) bool { return a >= v } }

func max(v int) func(int) bool { return func(a int) bool { return a < v } }

func TestNewFilterFromVS(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		a := assert.New(t, false)

		v := "str "
		f := NewFilterFromVS(StringPhrase("zero"), required[string], trimRight)("name", &v)
		name, msg := f()
		a.Empty(name).
			Nil(msg).
			Equal(v, "str")

		v = ""
		name, msg = f()
		a.Equal(name, "name").
			Equal(msg, StringPhrase("zero"))
	})

	t.Run("object", func(t *testing.T) {
		a := assert.New(t, false)

		v := object{Name: "name"}
		f := NewFilterFromVS(StringPhrase("zero"), required[object], func(t *object) { t.Name = "obj1" })("name", &v)
		name, msg := f()
		a.Empty(name).
			Nil(msg).
			Equal(v, object{Name: "obj1"})

		v = object{}
		f = NewFilterFromVS(StringPhrase("zero"), required[object])("name", &v)
		name, msg = f()
		a.Equal(name, "name").
			Equal(msg, StringPhrase("zero"))
	})
}

func TestNewFilter(t *testing.T) {
	a := assert.New(t, false)

	v := obj2{}
	f := NewFilter(nil, func(t *object) { t.Age = 18 })("name", &v.o2)
	name, msg := f()
	a.Empty(name).
		Nil(msg).
		Equal(v, obj2{o2: object{Age: 18}})

	v = obj2{}
	f = NewFilter(nil, func(t **object) { *t = &object{Name: "obj1"} })("name", &v.o1)
	name, msg = f()
	a.Empty(name).Nil(msg).
		Equal(v, obj2{o1: &object{Name: "obj1"}})
}

func TestNewSliceFilter(t *testing.T) {
	a := assert.New(t, false)
	message := StringPhrase("error")
	rule := NewRule(func(val int) bool { return val > 0 }, message)

	f := NewSliceFilter[int, []int](rule, func(t *int) { *t += 1 })
	a.NotNil(f)
	v := []int{1, 2, 3, 4, 5}
	name, msg := f("slice", &v)()
	a.Empty(name).Nil(msg).
		Equal(v, []int{2, 3, 4, 5, 6})

	// rule == nil
	f = NewSliceFilter[int, []int](nil, func(t *int) { *t += 1 })
	a.NotNil(f)
	v = []int{1, 2, 3, 4, 5}
	name, msg = f("slice", &v)()
	a.Empty(name).Nil(msg).
		Equal(v, []int{2, 3, 4, 5, 6})

	f = NewSliceFilter[int, []int](rule, func(t *int) { *t -= 1 })
	a.NotNil(f)
	v = []int{1, 2, 3, 4, 5}
	name, msg = f("slice", &v)()
	a.Equal(name, "slice[0]").
		Equal(msg, StringPhrase("error"))
}

func TestNewMapFilter(t *testing.T) {
	a := assert.New(t, false)
	message := StringPhrase("error")
	rule := NewRule[int](func(val int) bool { return val > 0 }, message)

	f := NewMapFilter[int, int, map[int]int](rule, func(t *int) { *t += 1 })
	a.NotNil(f)
	v := map[int]int{1: 1, 2: 2, 3: 3, 4: 4, 5: 5}
	name, msg := f("map", &v)()
	a.Empty(name).Nil(msg).
		Equal(v, map[int]int{1: 2, 2: 3, 3: 4, 4: 5, 5: 6})

	// rule == nil
	f = NewMapFilter[int, int, map[int]int](nil, func(t *int) { *t += 1 })
	a.NotNil(f)
	v = map[int]int{1: 1, 2: 2, 3: 3, 4: 4, 5: 5}
	name, msg = f("map", &v)()
	a.Empty(name).Nil(msg).
		Equal(v, map[int]int{1: 2, 2: 3, 3: 4, 4: 5, 5: 6})

	f = NewMapFilter[int, int, map[int]int](rule, func(t *int) { *t -= 1 })
	a.NotNil(f)
	v = map[int]int{1: 1, 2: 2, 3: 3, 4: 4, 5: 5}
	name, msg = f("map", &v)()
	a.Equal(name, "map[1]").
		Equal(msg, StringPhrase("error"))
}

func TestRulerOf(t *testing.T) {
	a := assert.New(t, false)

	r1 := NewRule(zero[int], StringPhrase("r1"))
	name, msg := r1("r1", 5)
	a.Equal(name, "r1").Equal(msg, StringPhrase("r1"))
	name, msg = r1("r1", 0)
	a.Empty(name).Nil(msg)

	r2 := NewRule(between(-1, 50), StringPhrase("r2"))
	name, msg = r2("r2", 5)
	a.Empty(name).Nil(msg)

	t.Run("NewRules", func(t *testing.T) {
		rs := NewRules(r1, r2)
		name, msg = rs("rs", 5)
		a.Equal(name, "rs").Equal(msg, StringPhrase("r1"))

		name, msg = rs("rs", 0)
		a.Empty(name).Nil(msg)

		name, msg = rs("rs", -2)
		a.Equal(name, "rs").Equal(msg, StringPhrase("r1"))
	})

	t.Run("NewSliceRule", func(t *testing.T) {
		rs := NewSliceRule[int, []int](between(1, 5), StringPhrase("rs"))

		name, msg = rs("rs", []int{1, 2, 3})
		a.Empty(name).Nil(msg)

		name, msg = rs("rs", []int{1, 2, 3, 7})
		a.Equal(name, "rs[3]").Equal(msg, StringPhrase("rs"))
	})

	t.Run("NewSliceRules", func(t *testing.T) {
		rs := NewSliceRules[int, []int](r1, r2)

		name, msg = rs("rs", []int{0})
		a.Empty(name).Nil(msg)

		name, msg = rs("rs", []int{-1, 0, 1})
		a.Equal(name, "rs[0]").Equal(msg, StringPhrase("r1"))
	})

	t.Run("NewMapRule", func(t *testing.T) {
		rm := NewMapRule[string, int, map[string]int](between(1, 5), StringPhrase("rm"))

		name, msg = rm("rm", map[string]int{"1": 1, "2": 2})
		a.Empty(name).Nil(msg)

		name, msg = rm("rm", map[string]int{"1": 1, "err": 7, "2": 2})
		a.Equal(name, "rm[err]").Equal(msg, StringPhrase("rm"))
	})

	t.Run("NewMapRules", func(t *testing.T) {
		rm := NewMapRules[string, int, map[string]int](r1, r2)

		name, msg = rm("rm", map[string]int{"0": 0})
		a.Empty(name).Nil(msg)

		name, msg = rm("rm", map[string]int{"0": 0, "1": 1})
		a.Equal(name, "rm[1]").Equal(msg, StringPhrase("r1"))
	})
}

func TestFilterContext(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	ctx := s.NewContext(w, r)

	min_2 := NewRule(min(-2), Phrase("-2"))
	min_3 := NewRule(min(-3), Phrase("-3"))
	max50 := NewRule(max(50), Phrase("50"))
	max_4 := NewRule(max(-4), Phrase("-4"))

	n100 := -100
	p100 := 100
	v := ctx.newFilterContext(false).
		Add(NewFilter(NewRules(min_2, min_3))("f1", &n100)).
		Add(NewFilter(NewRules(max50, max_4))("f2", &p100))
	a.Equal(v.problem.Params, []RFC7807Param{
		{Name: "f1", Reason: "-2"},
		{Name: "f2", Reason: "50"},
	})

	n100 = -100
	p100 = 100
	v = ctx.newFilterContext(true).
		Add(NewFilter(NewRules(min_2, min_3))("f1", &n100)).
		Add(NewFilter(NewRules(max50, max_4))("f2", &p100))
	a.Equal(v.problem.Params, []RFC7807Param{
		{Name: "f1", Reason: "-2"},
	})
}

func TestFilterContext_New(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	ctx := s.NewContext(w, r)

	v := ctx.newFilterContext(false)
	v1 := v.New("v1.", func(f *FilterContext) {
		f.AddReason("f1", StringPhrase("s1"))
		v2 := f.New("v2.", func(f *FilterContext) {
			f.AddError("f2", errors.New("s2"))
		})
		a.Equal(v2, f)
	})
	a.Equal(v1, v)

	a.Equal(v.problem.Params, []RFC7807Param{
		{Name: "v1.f1", Reason: "s1"},
		{Name: "v1.v2.f2", Reason: "s2"},
	})
}

func TestFilterContext_When(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	ctx := s.NewContext(w, r)

	min18 := NewRule(min(18), Phrase("不能小于 18"))
	notEmpty := NewRule(required[string], Phrase("不能为空"))

	obj := &object{}
	v := ctx.newFilterContext(false).
		Add(NewFilter(min18)("obj/age", &obj.Age)).
		When(obj.Age > 0, func(v *FilterContext) {
			v.Add(NewFilter(notEmpty)("obj/name", &obj.Name))
		})
	a.Equal(v.problem.Params, []RFC7807Param{
		{Name: "obj/age", Reason: "不能小于 18"},
	})

	obj = &object{Age: 15}
	v = ctx.newFilterContext(false).
		Add(NewFilter(min18)("obj/age", &obj.Age)).
		When(obj.Age > 0, func(v *FilterContext) {
			v.Add(NewFilter(notEmpty)("obj/name", &obj.Name))
		})
	a.Equal(v.problem.Params, []RFC7807Param{
		{Name: "obj/age", Reason: "不能小于 18"},
		{Name: "obj/name", Reason: "不能为空"},
	})
}
