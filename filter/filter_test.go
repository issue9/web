// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
	"errors"
	"testing"

	"github.com/issue9/assert/v6"
	"github.com/issue9/localeutil"
)

type object struct {
	Name string `json:"name"`
	Age  int
}

type collector struct {
	params map[string]localeutil.Stringer
}

func newCollector() *collector { return &collector{params: make(map[string]localeutil.Stringer, 3)} }

func (c *collector) collect(name string, reason localeutil.Stringer) {
	c.params[name] = reason
}

func TestFilter(t *testing.T) {
	a := assert.New(t, false)

	min2 := ValidatorRule(buildMinValidator(-2), localeutil.Phrase("-2"))
	min3 := ValidatorRule(buildMinValidator(-3), localeutil.Phrase("-3"))
	max50 := ValidatorRule(buildMaxValidator(50), localeutil.Phrase("50"))
	max4 := ValidatorRule(buildMaxValidator(-4), localeutil.Phrase("-4"))

	f, c := newFilter(false)
	n100 := -100
	p100 := 100
	f.Add("f1", &n100, min2, min3).Add("f2", &p100, max50, max4)
	a.Equal(c.params, map[string]localeutil.Stringer{
		"f1": localeutil.Phrase("-2"),
		"f2": localeutil.Phrase("50"),
	})

	f, c = newFilter(true)
	n100 = -100
	p100 = 100
	f.Add("f1", &n100, min2, min3).Add("f2", &p100, max50, max4)
	a.Equal(c.params, map[string]localeutil.Stringer{
		"f1": localeutil.Phrase("-2"),
	})
}

func TestFilter_AddFilter(t *testing.T) {
	a := assert.New(t, false)

	f, c := newFilter(false)
	f1 := f.AddFilter("v1.", func(f *Filter) {
		f.AddReason("f1", localeutil.StringPhrase("s1"))
		f2 := f.AddFilter("v2.", func(f *Filter) {
			f.AddError("f2", errors.New("s2"))
		})
		a.Equal(f2, f)
	})
	a.Equal(f1, f)

	a.Equal(c.params, map[string]localeutil.Stringer{
		"v1.f1":    localeutil.StringPhrase("s1"),
		"v1.v2.f2": localeutil.StringPhrase("s2"),
	}).True(f.HasError())
}

func TestFilter_When(t *testing.T) {
	a := assert.New(t, false)

	min18 := ValidatorRule(buildMinValidator(18), localeutil.Phrase("不能小于 18"))
	notEmpty := ValidatorRule(required[string], localeutil.Phrase("不能为空"))

	f, c := newFilter(false)
	obj := &object{}
	f.Add("obj/age", &obj.Age, min18).
		When(obj.Age > 0, func(v *Filter) {
			v.Add("obj/name", &obj.Name, notEmpty)
		})
	a.Equal(c.params, map[string]localeutil.Stringer{
		"obj/age": localeutil.Phrase("不能小于 18"),
	})

	f, c = newFilter(false)
	obj = &object{Age: 15}
	f.Add("obj/age", &obj.Age, min18).
		When(obj.Age > 0, func(v *Filter) {
			v.Add("obj/name", &obj.Name, notEmpty)
		})
	a.Equal(c.params, map[string]localeutil.Stringer{
		"obj/age":  localeutil.Phrase("不能小于 18"),
		"obj/name": localeutil.Phrase("不能为空"),
	})
}
