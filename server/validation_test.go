// SPDX-License-Identifier: MIT

package server

import (
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/assert/v3/rest"
	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"
)

type (
	root1 struct {
		Root *root2
		F1   int
	}

	root2 struct {
		O1 *object
		O2 *object
	}

	object struct {
		Name string
		Age  int
	}
)

func required(v any) bool { return !reflect.ValueOf(v).IsZero() }

func min(v int) ValidateFunc {
	return func(a any) bool {
		if num, ok := a.(int); ok {
			return num >= v
		}
		return false
	}
}

func max(v int) ValidateFunc {
	return func(a any) bool {
		if num, ok := a.(int); ok {
			return num < v
		}
		return false
	}
}

func in(element ...int) ValidateFunc {
	return func(v any) bool {
		return sliceutil.Exists(element, func(elem int) bool { return elem == v })
	}
}

func notIn(element ...int) ValidateFunc {
	return func(v any) bool {
		return !sliceutil.Exists(element, func(elem int) bool { return elem == v })
	}
}

func TestContext_newValidation(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, nil)
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	ctx := s.newContext(w, r, nil)

	min_2 := NewRule(localeutil.Phrase("-2"), min(-2))
	min_3 := NewRule(localeutil.Phrase("-3"), min(-3))
	max50 := NewRule(localeutil.Phrase("50"), max(50))
	max_4 := NewRule(localeutil.Phrase("-4"), max(-4))

	v := ctx.newValidation(false).
		AddField(-100, "f1", min_2, min_3).
		AddField(100, "f2", max50, max_4)
	a.Equal(v.keys, []string{"f1", "f2"}).
		Equal(v.reasons, []string{"-2", "50"})

	v = ctx.newValidation(true).
		AddField(-100, "f1", min_2, min_3).
		AddField(100, "f2", max50, max_4)
	a.Equal(v.keys, []string{"f1"}).
		Equal(v.reasons, []string{"-2"})
}

func TestValidation_AddField(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, nil)
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	ctx := s.newContext(w, r, nil)

	min18 := NewRule(localeutil.Phrase("不能小于 18"), min(18))
	min5 := NewRule(localeutil.Phrase("min-5"), min(5))

	obj := &object{}
	v := ctx.newValidation(false).
		AddField(obj.Age, "obj/age", min18)
	a.Equal(v.keys, []string{"obj/age"}).
		Equal(v.reasons, []string{"不能小于 18"})

	// object
	root := root2{}
	v = ctx.newValidation(false)
	v.AddField(root.O1, "o1", NewRuleFunc(localeutil.Phrase("o1 required"), required)).
		AddField(root.O2, "o2", NewRuleFunc(localeutil.Phrase("o2 required"), required))
	a.Equal(v.keys, []string{"o1", "o2"}).
		Equal(v.reasons, []string{"o1 required", "o2 required"})

	root = root2{O1: &object{}}
	v = ctx.newValidation(false)
	v.AddField(root.O1.Age, "o1.age", min18)
	a.Equal(v.keys, []string{"o1.age"}).
		Equal(v.reasons, []string{"不能小于 18"})

	v = ctx.newValidation(false)
	rv := root1{Root: &root2{O1: &object{}}}
	v.AddField(rv.Root.O1.Age, "root/o1/age", min18).
		AddField(rv.F1, "f1", min5)
	a.Equal(v.keys, []string{"root/o1/age", "f1"}).
		Equal(v.reasons, []string{"不能小于 18", "min-5"})
}

func TestValidation_AddSliceField(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, nil)
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	ctx := s.newContext(w, r, nil)

	min5 := NewRule(localeutil.Phrase("min-5"), min(5))

	// 将数组当普通元素处理
	v := ctx.newValidation(false).
		AddField([]int{1, 2, 6}, "slice", min5)
	a.Equal(v.keys, []string{"slice"}).
		Equal(v.reasons, []string{"min-5"})

	// 普通元素指定为 slice
	v = ctx.newValidation(false).
		AddSliceField(123456, "slice", min5)
	a.Equal(v.keys, []string{"slice"}).
		Equal(v.reasons, []string{isNotSlice.LocaleString(ctx.LocalePrinter())})

	v = ctx.newValidation(true).
		AddSliceField(123456, "slice", min5)
	a.Equal(v.keys, []string{"slice"}).
		Equal(v.reasons, []string{isNotSlice.LocaleString(ctx.LocalePrinter())})

	// exitAtError = false
	v = ctx.newValidation(false).
		AddSliceField([]int{1, 2, 6}, "slice", min5)
	a.Equal(v.keys, []string{"slice[0]", "slice[1]"}).
		Equal(v.reasons, []string{"min-5", "min-5"})

	// exitAtError = true
	v = ctx.newValidation(true).
		AddSliceField([]int{1, 2, 6}, "slice", min5)
	a.Equal(v.keys, []string{"slice[0]"}).
		Equal(v.reasons, []string{"min-5"})
}

func TestValidation_AddMapField(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, nil)
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	ctx := s.newContext(w, r, nil)

	min5 := NewRule(localeutil.Phrase("min-5"), min(5))

	// 将数组当普通元素处理
	v := ctx.newValidation(false).
		AddField([]int{1, 2, 6}, "map", min5)
	a.Equal(v.keys, []string{"map"}).
		Equal(v.reasons, []string{"min-5"})

	// 普通元素指定为 map
	v = ctx.newValidation(false).
		AddMapField(123456, "map", min5)
	a.Equal(v.keys, []string{"map"}).
		Equal(v.reasons, []string{isNotMap.LocaleString(ctx.LocalePrinter())})

	v = ctx.newValidation(true).
		AddMapField(123456, "map", min5)
	a.Equal(v.keys, []string{"map"}).
		Equal(v.reasons, []string{isNotMap.LocaleString(ctx.LocalePrinter())})

	// exitAtError = false
	v = ctx.newValidation(false).
		AddMapField(map[string]int{"0": 1, "2": 2, "6": 6}, "map", min5)
	sort.Strings(v.keys) // NOTE: 排序会破坏 v.keys 和 v.reasons 之间的关联。
	a.Equal(v.keys, []string{"map[0]", "map[2]"}).
		Equal(v.reasons, []string{"min-5", "min-5"})

	// exitAtError = true
	v = ctx.newValidation(true).
		AddMapField(map[string]int{"0": 1, "2": 2, "6": 6}, "map", min5)
	a.Length(v.keys, 1) // map 顺序未定，无法确定哪个元素会被称验证
}

func TestValidation_When(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, nil)
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	ctx := s.newContext(w, r, nil)

	min18 := NewRule(localeutil.Phrase("不能小于 18"), min(18))
	notEmpty := NewRule(localeutil.Phrase("不能为空"), ValidateFunc(required))

	obj := &object{}
	v := ctx.newValidation(false).
		AddField(obj, "obj/age", min18).
		When(obj.Age > 0, func(v *Validation) {
			v.AddField(obj.Name, "obj/name", notEmpty)
		})
	a.Equal(v.keys, []string{"obj/age"}).
		Equal(v.reasons, []string{"不能小于 18"})

	obj = &object{Age: 15}
	v = ctx.newValidation(false).
		AddField(obj, "obj/age", min18).
		When(obj.Age > 0, func(v *Validation) {
			v.AddField(obj.Name, "obj/name", notEmpty)
		})
	a.Equal(v.keys, []string{"obj/age", "obj/name"}).
		Equal(v.reasons, []string{"不能小于 18", "不能为空"})
}

func TestAndValidator(t *testing.T) {
	a := assert.New(t, false)

	v := AndValidator(in(1, 2, 3), notIn(2, 3, 4))
	a.True(v.IsValid(1))
	a.False(v.IsValid(2))
	a.False(v.IsValid(-1))
	a.False(v.IsValid(100))
}

func TestOrValidator(t *testing.T) {
	a := assert.New(t, false)

	v := OrValidator(in(1, 2, 3), notIn(2, 3, 4))
	a.True(v.IsValid(1))
	a.True(v.IsValid(2))
	a.False(v.IsValid(4))
	a.True(v.IsValid(-1))
	a.True(v.IsValid(100))
}

func TestAndValidateFunc(t *testing.T) {
	a := assert.New(t, false)

	v := AndValidateFunc(in(1, 2, 3).IsValid, notIn(2, 3, 4).IsValid)
	a.True(v.IsValid(1))
	a.False(v.IsValid(2))
	a.False(v.IsValid(-1))
	a.False(v.IsValid(100))
}

func TestOrValidateFunc(t *testing.T) {
	a := assert.New(t, false)

	v := OrValidateFunc(in(1, 2, 3).IsValid, notIn(2, 3, 4).IsValid)
	a.True(v.IsValid(1))
	a.True(v.IsValid(2))
	a.False(v.IsValid(4))
	a.True(v.IsValid(-1))
	a.True(v.IsValid(100))
}
