// SPDX-License-Identifier: MIT

package filter

import (
	"reflect"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
)

func zero[T any](v T) bool { return reflect.ValueOf(v).IsZero() }

func between[T ~int | ~uint | float32 | float64](min, max T) ValidatorFuncOf[T] {
	return ValidatorFuncOf[T](func(vv T) bool { return vv >= min && vv <= max })
}

func TestNot(t *testing.T) {
	a := assert.New(t, false)

	z := zero[int]
	a.True(z(0)).False(z(1))

	nz := Not(zero[int])
	a.False(nz(0)).True(nz(1))
}

func TestAnd_OrFunc(t *testing.T) {
	a := assert.New(t, false)

	and := And(between(0, 100), between(-1, 50))
	a.True(and(0)).
		True(and(1)).
		False(and(51))

	or := Or(between(0, 100), between(-1, 50))
	a.True(or(0)).
		True(or(1)).
		False(or(500))
}

func TestRulerOf(t *testing.T) {
	a := assert.New(t, false)

	r1 := NewRule(zero[int], localeutil.Phrase("r1"))
	name, msg := r1("r1", 5)
	a.Equal(name, "r1").Equal(msg, localeutil.Phrase("r1"))
	name, msg = r1("r1", 0)
	a.Empty(name).Nil(msg)

	r2 := NewRule(between(-1, 50), localeutil.Phrase("r2"))
	name, msg = r2("r2", 5)
	a.Empty(name).Nil(msg)

	t.Run("NewRules", func(t *testing.T) {
		rs := NewRules(r1, r2)
		name, msg = rs("rs", 5)
		a.Equal(name, "rs").Equal(msg, localeutil.Phrase("r1"))

		name, msg = rs("rs", 0)
		a.Empty(name).Nil(msg)

		name, msg = rs("rs", -2)
		a.Equal(name, "rs").Equal(msg, localeutil.Phrase("r1"))
	})

	t.Run("NewSliceRule", func(t *testing.T) {
		rs := NewSliceRule[int, []int](between(1, 5), localeutil.Phrase("rs"))

		name, msg = rs("rs", []int{1, 2, 3})
		a.Empty(name).Nil(msg)

		name, msg = rs("rs", []int{1, 2, 3, 7})
		a.Equal(name, "rs[3]").Equal(msg, localeutil.Phrase("rs"))
	})

	t.Run("NewSliceRules", func(t *testing.T) {
		rs := NewSliceRules[int, []int](r1, r2)

		name, msg = rs("rs", []int{0})
		a.Empty(name).Nil(msg)

		name, msg = rs("rs", []int{-1, 0, 1})
		a.Equal(name, "rs[0]").Equal(msg, localeutil.Phrase("r1"))
	})

	t.Run("NewMapRule", func(t *testing.T) {
		rm := NewMapRule[string, int, map[string]int](between(1, 5), localeutil.Phrase("rm"))

		name, msg = rm("rm", map[string]int{"1": 1, "2": 2})
		a.Empty(name).Nil(msg)

		name, msg = rm("rm", map[string]int{"1": 1, "err": 7, "2": 2})
		a.Equal(name, "rm[err]").Equal(msg, localeutil.Phrase("rm"))
	})

	t.Run("NewMapRules", func(t *testing.T) {
		rm := NewMapRules[string, int, map[string]int](r1, r2)

		name, msg = rm("rm", map[string]int{"0": 0})
		a.Empty(name).Nil(msg)

		name, msg = rm("rm", map[string]int{"0": 0, "1": 1})
		a.Equal(name, "rm[1]").Equal(msg, localeutil.Phrase("r1"))
	})
}
