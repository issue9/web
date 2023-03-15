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

	"github.com/issue9/web/validation"
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

func required[T any](v T) bool { return !reflect.ValueOf(v).IsZero() }

func min(v int) validation.ValidatorOf[int] {
	return validation.ValidatorFuncOf[int](func(a int) bool { return a >= v })
}

func max(v int) validation.ValidatorOf[int] {
	return validation.ValidatorFuncOf[int](func(a int) bool { return a < v })
}

func in(element ...int) validation.ValidatorFuncOf[int] {
	return func(v int) bool {
		return sliceutil.Exists(element, func(elem int) bool { return elem == v })
	}
}

func notIn(element ...int) validation.ValidatorFuncOf[int] {
	return func(v int) bool {
		return !sliceutil.Exists(element, func(elem int) bool { return elem == v })
	}
}

func TestContext_NewValidation(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a, nil)
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	ctx := s.newContext(w, r, nil)

	min_2 := validation.NewRuleOf(min(-2), localeutil.Phrase("-2"))
	min_3 := validation.NewRuleOf(min(-3), localeutil.Phrase("-3"))
	max50 := validation.NewRuleOf(max(50), localeutil.Phrase("50"))
	max_4 := validation.NewRuleOf(max(-4), localeutil.Phrase("-4"))

	v := ctx.NewValidation(false).
		AddField(validation.NewRulesOf(min_2, min_3).Build("f1", -100)).
		AddField(validation.NewRulesOf(max50, max_4).Build("f2", 100))
	a.Equal(v.keys, []string{"f1", "f2"}).
		Equal(v.reasons, []string{"-2", "50"})

	v = ctx.NewValidation(true).
		AddField(validation.NewRulesOf(min_2, min_3).Build("f1", -100)).
		AddField(validation.NewRulesOf(max50, max_4).Build("f2", 100))
	a.Equal(v.keys, []string{"f1"}).
		Equal(v.reasons, []string{"-2"})

	// object

	root := root2{}
	req := validation.ValidatorFuncOf[*object](required[*object])
	v = ctx.NewValidation(false)
	v.AddField(validation.NewRuleOf[*object](req, localeutil.Phrase("o1 required")).Build("o1", root.O1)).
		AddField(validation.NewRuleOf[*object](req, localeutil.Phrase("o2 required")).Build("o2", root.O2))
	a.Equal(v.keys, []string{"o1", "o2"}).
		Equal(v.reasons, []string{"o1 required", "o2 required"})

	min18 := validation.NewRuleOf(min(18), localeutil.Phrase("不能小于 18"))
	min5 := validation.NewRuleOf(min(5), localeutil.Phrase("min-5"))

	root = root2{O1: &object{}}
	v = ctx.NewValidation(false)
	v.AddField(min18.Build("o1.age", root.O1.Age))
	a.Equal(v.keys, []string{"o1.age"}).
		Equal(v.reasons, []string{"不能小于 18"})

	v = ctx.NewValidation(false)

	rv := root1{Root: &root2{O1: &object{}}}
	v.AddField(min18.Build("root/o1/age", rv.Root.O1.Age)).
		AddField(min5.Build("f1", rv.F1))
	a.Equal(v.keys, []string{"root/o1/age", "f1"}).
		Equal(v.reasons, []string{"不能小于 18", "min-5"})
}

func TestValidation_When(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a, nil)
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	ctx := s.newContext(w, r, nil)

	min18 := validation.NewRuleOf(min(18), localeutil.Phrase("不能小于 18"))
	req := validation.ValidatorFuncOf[string](required[string])
	notEmpty := validation.NewRuleOf[string](req, localeutil.Phrase("不能为空"))

	obj := &object{}
	v := ctx.NewValidation(false).
		AddField(min18.Build("obj/age", obj.Age)).
		When(obj.Age > 0, func(v *Validation) {
			v.AddField(notEmpty.Build("obj/name", obj.Name))
		})
	a.Equal(v.keys, []string{"obj/age"}).
		Equal(v.reasons, []string{"不能小于 18"})

	obj = &object{Age: 15}
	v = ctx.NewValidation(false).
		AddField(min18.Build("obj/age", obj.Age)).
		When(obj.Age > 0, func(v *Validation) {
			v.AddField(notEmpty.Build("obj/name", obj.Name))
		})
	a.Equal(v.keys, []string{"obj/age", "obj/name"}).
		Equal(v.reasons, []string{"不能小于 18", "不能为空"})
}
