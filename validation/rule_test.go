// SPDX-License-Identifier: MIT

package validation

import (
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
)

func TestRulerOf(t *testing.T) {
	a := assert.New(t, false)

	r1 := NewRuleOf[int](ValidatorFuncOf[int](zero[int]), localeutil.Phrase("r1"))
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

	t.Run("NewSliceRuleOf", func(t *testing.T) {
		v := ValidatorFuncOf[int](between(1, 5))
		rs := NewSliceRuleOf[int, []int](v, localeutil.Phrase("rs"))

		name, msg = rs.Build("rs", []int{1, 2, 3}).Validate()
		a.Empty(name).Nil(msg)

		name, msg = rs.Build("rs", []int{1, 2, 3, 7}).Validate()
		a.Equal(name, "rs[3]").Equal(msg, localeutil.Phrase("rs"))
	})

	t.Run("NewMapRuleOf", func(t *testing.T) {
		v := ValidatorFuncOf[int](between(1, 5))
		rm := NewMapRuleOf[string, int, map[string]int](v, localeutil.Phrase("rm"))

		name, msg = rm.Build("rm", map[string]int{"1": 1, "2": 2}).Validate()
		a.Empty(name).Nil(msg)

		name, msg = rm.Build("rm", map[string]int{"1": 1, "err": 7, "2": 2}).Validate()
		a.Equal(name, "rm[err]").Equal(msg, localeutil.Phrase("rm"))
	})
}
