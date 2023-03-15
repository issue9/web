// SPDX-License-Identifier: MIT

package validation

import (
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
)

func TestRulerOf(t *testing.T) {
	a := assert.New(t, false)

	r1 := NewRuleFuncOf(zero[int], localeutil.Phrase("r1"))
	name, msg := r1.Build("r1", 5).Validate()
	a.Equal(name, "r1").Equal(msg, localeutil.Phrase("r1"))
	name, msg = r1.Build("r1", 0).Validate()
	a.Empty(name).Nil(msg)

	r2 := NewRuleOf[int](between(-1, 50), localeutil.Phrase("r2"))
	name, msg = r2.Build("r2", 5).Validate()
	a.Empty(name).Nil(msg)

	t.Run("NewRulesOf", func(t *testing.T) {
		rs := NewRulesOf(r1, r2)
		name, msg = rs.Build("rs", 5).Validate()
		a.Equal(name, "rs").Equal(msg, localeutil.Phrase("r1"))

		name, msg = rs.Build("rs", 0).Validate()
		a.Empty(name).Nil(msg)

		name, msg = rs.Build("rs", -2).Validate()
		a.Equal(name, "rs").Equal(msg, localeutil.Phrase("r1"))
	})

	t.Run("NewSliceRuleFuncOf", func(t *testing.T) {
		rs := NewSliceRuleFuncOf[int, []int](between(1, 5), localeutil.Phrase("rs"))

		name, msg = rs.Build("rs", []int{1, 2, 3}).Validate()
		a.Empty(name).Nil(msg)

		name, msg = rs.Build("rs", []int{1, 2, 3, 7}).Validate()
		a.Equal(name, "rs[3]").Equal(msg, localeutil.Phrase("rs"))
	})

	t.Run("NewSliceRulesOf", func(t *testing.T) {
		rs := NewSliceRulesOf[int, []int](r1, r2)

		name, msg = rs.Build("rs", []int{0}).Validate()
		a.Empty(name).Nil(msg)

		name, msg = rs.Build("rs", []int{-1, 0, 1}).Validate()
		a.Equal(name, "rs[0]").Equal(msg, localeutil.Phrase("r1"))
	})

	t.Run("NewMapRuleFuncOf", func(t *testing.T) {
		rm := NewMapRuleFuncOf[string, int, map[string]int](between(1, 5), localeutil.Phrase("rm"))

		name, msg = rm.Build("rm", map[string]int{"1": 1, "2": 2}).Validate()
		a.Empty(name).Nil(msg)

		name, msg = rm.Build("rm", map[string]int{"1": 1, "err": 7, "2": 2}).Validate()
		a.Equal(name, "rm[err]").Equal(msg, localeutil.Phrase("rm"))
	})

	t.Run("NewMapRulesOf", func(t *testing.T) {
		rm := NewMapRulesOf[string, int, map[string]int](r1, r2)

		name, msg = rm.Build("rm", map[string]int{"0": 0}).Validate()
		a.Empty(name).Nil(msg)

		name, msg = rm.Build("rm", map[string]int{"0": 0, "1": 1}).Validate()
		a.Equal(name, "rm[1]").Equal(msg, localeutil.Phrase("r1"))
	})
}
