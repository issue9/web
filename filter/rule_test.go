// SPDX-License-Identifier: MIT

package filter

import (
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
)

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
