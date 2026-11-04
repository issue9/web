// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
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

func TestFilter(t *testing.T) {
	a := assert.New(t, false)

	min2 := V2R(localeutil.Phrase("-2"), validators.GreatEqual(-2))
	min3 := V2R(localeutil.Phrase("-3"), validators.GreatEqual(-3))
	max50 := V2R(localeutil.Phrase("50"), validators.Less(50))
	max4 := V2R(localeutil.Phrase("-4"), validators.Less(-4))

	f, c := newFilter(false, "")
	n100 := -100
	p100 := 100
	f.ApplyRule("f1", &n100, min2, min3).ApplyRule("f2", &p100, max50, max4)
	a.Equal(c, map[string][]localeutil.Stringer{
		"f1": {localeutil.Phrase("-2")},
		"f2": {localeutil.Phrase("50")},
	})

	f, c = newFilter(true, "")
	n100 = -100
	p100 = 100
	f.ApplyRule("f1", &n100, min2, min3).ApplyRule("f2", &p100, max50, max4)
	a.Equal(c, map[string][]localeutil.Stringer{
		"f1": {localeutil.Phrase("-2")},
	})

	_, collector := CreateMapCollector()
	f.Reset("reset", false, collector)
	a.Equal(f.Name(), "reset").False(f.ExitOnError())
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
	f.ApplyRule("obj/age", &obj.Age, min18).
		When(func() bool { return obj.Age > 0 }, func(v *Filter) {
			v.ApplyRule("obj/name", &obj.Name, notEmpty)
		})
	a.Equal(c, map[string][]localeutil.Stringer{
		"obj/age": {localeutil.Phrase("不能小于 18")},
	})

	f, c = newFilter(false, "")
	obj = &object{Age: 15}
	f.ApplyRule("obj/age", &obj.Age, min18).
		When(func() bool { return obj.Age > 0 }, func(v *Filter) {
			v.ApplyRule("obj/name", &obj.Name, notEmpty)
		})
	a.Equal(c, map[string][]localeutil.Stringer{
		"obj/age":  {localeutil.Phrase("不能小于 18")},
		"obj/name": {localeutil.Phrase("不能为空")},
	})
}
