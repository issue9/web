// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
	"errors"
	"testing"

	"github.com/issue9/assert/v6"
	"github.com/issue9/localeutil"

	"github.com/issue9/web/filter/validators"
)

type object struct {
	Name string `json:"name"`
	Age  int
}

func newFilter(exitOnError bool, name string) (*Filter, Map) {
	m, c := CreateMapCollector()
	return New(name, exitOnError, c), m
}

func TestFilter_ApplyFilter(t *testing.T) {
	a := assert.New(t, false)

	f, c := newFilter(false, "")
	f1 := f.ApplyFilter("v1", func(f *Filter) {
		f.AddReason("f1", localeutil.StringPhrase("s1"))
		f2 := f.ApplyFilter("", func(f *Filter) {
			f.AddError("f2", errors.New("s2"))
		})
		a.Equal(f2, f)
	})
	a.Equal(f1, f)

	a.Equal(c, map[string][]localeutil.Stringer{
		"v1.f1": {localeutil.StringPhrase("s1")},
		"v1.f2": {localeutil.StringPhrase("s2")},
	}).True(f.HasError())
}

func TestFilter_AddReason(t *testing.T) {
	a := assert.New(t, false)

	f, c := newFilter(true, "")
	a.Panic(func() { f.AddReason("name", nil) })
	a.Empty(c)

	f, c = newFilter(true, "")
	f.AddReason("name", localeutil.Phrase("reason"))
	a.NotEmpty(c["name"])

	f, c = newFilter(true, "")
	a.Panic(func() { f.AddReason("", localeutil.Phrase("reason")) })
	a.Empty(c)

	f, c = newFilter(true, "parent")
	f.AddReason("name", localeutil.Phrase("reason"))
	a.NotEmpty(c["parent.name"])

	f, c = newFilter(true, "parent")
	f.AddReason("", localeutil.Phrase("reason"))
	a.NotEmpty(c["parent"])

	f.AddReason("name", localeutil.Phrase("reason"))
	a.NotEmpty(c["parent"]).Empty(c["parent.name"])
}

func TestFilter_When(t *testing.T) {
	a := assert.New(t, false)

	min18 := V2R(localeutil.Phrase("不能小于 18"), validators.GreatEqual(18))
	notEmpty := V2R(localeutil.Phrase("不能为空"), required[string])

	f, c := newFilter(false, "")
	obj := &object{}
	f.Apply("obj/age", &obj.Age, min18).
		When(func() bool { return obj.Age > 0 }, func(v *Filter) {
			v.Apply("obj/name", &obj.Name, notEmpty)
		})
	a.Equal(c, map[string][]localeutil.Stringer{
		"obj/age": {localeutil.Phrase("不能小于 18")},
	})

	f, c = newFilter(false, "")
	obj = &object{Age: 15}
	f.Apply("obj/age", &obj.Age, min18).
		When(func() bool { return obj.Age > 0 }, func(v *Filter) {
			v.Apply("obj/name", &obj.Name, notEmpty)
		})
	a.Equal(c, map[string][]localeutil.Stringer{
		"obj/age":  {localeutil.Phrase("不能小于 18")},
		"obj/name": {localeutil.Phrase("不能为空")},
	})
}
