// SPDX-FileCopyrightText: 2024-2026 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
	"testing"

	"github.com/issue9/assert/v6"
	"github.com/issue9/localeutil"

	"github.com/issue9/web/filter/sanitizers"
	"github.com/issue9/web/filter/validators"
)

func required[T any](v T) bool { return validators.Not[T](validators.ZeroAny)(v) }

func TestV2R(t *testing.T) {
	a := assert.New(t, false)

	id := " "
	v, c := newFilter(false, "")
	v.Apply("id", &id, V2R[string](localeutil.StringPhrase("required"), required))
	a.Empty(c).Equal(id, " ")

	id = "X "
	v, c = newFilter(false, "")
	v.Apply("id", &id, S2R(sanitizers.Upper), V2R[string](localeutil.StringPhrase("required"), required))
	a.Empty(c).Equal(id, "X ")
}

func TestFilter_Valid(t *testing.T) {
	t.Run("验证通过", func(t *testing.T) {
		a := assert.New(t, false)
		f, c := newFilter(false, "parent")
		f.Valid("age", 20, validators.GreatEqual(18), localeutil.Phrase("too young"))
		a.Empty(c)
	})

	t.Run("验证失败", func(t *testing.T) {
		a := assert.New(t, false)
		f, c := newFilter(false, "parent")
		f.Valid("age", 16, validators.GreatEqual(18), localeutil.Phrase("too young"))
		a.Equal(c, map[string][]localeutil.Stringer{
			"parent.age": {localeutil.Phrase("too young")},
		})
	})

	t.Run("已有错误时提前退出", func(t *testing.T) {
		a := assert.New(t, false)
		f, c := newFilter(true, "")
		f.AddReason("existing", localeutil.Phrase("existing error"))
		called := false
		f.Valid("age", 16, func(int) bool {
			called = true
			return false
		}, localeutil.Phrase("too young"))
		a.False(called).Equal(c, map[string][]localeutil.Stringer{
			"existing": {localeutil.Phrase("existing error")},
		})
	})
}

func TestFilter_ValidSlice(t *testing.T) {
	t.Run("全部通过", func(t *testing.T) {
		a := assert.New(t, false)
		f, c := newFilter(false, "")
		f.ValidSlice("values", []int{1, 2, 3}, validators.Great(0), localeutil.Phrase("invalid"))
		a.Empty(c)
	})

	t.Run("收集所有错误", func(t *testing.T) {
		a := assert.New(t, false)
		f, c := newFilter(false, "")
		f.ValidSlice("values", []int{1, 0, -1}, validators.Great(0), localeutil.Phrase("invalid"))
		a.Equal(c, map[string][]localeutil.Stringer{
			"values[1]": {localeutil.Phrase("invalid")},
			"values[2]": {localeutil.Phrase("invalid")},
		})
	})

	t.Run("遇错后停止", func(t *testing.T) {
		a := assert.New(t, false)
		f, c := newFilter(true, "")
		f.ValidSlice("values", []int{0, -1}, validators.Great(0), localeutil.Phrase("invalid"))
		a.Equal(c, map[string][]localeutil.Stringer{
			"values[0]": {localeutil.Phrase("invalid")},
		})
	})
}

func TestFilter_ValidMap(t *testing.T) {
	t.Run("全部通过", func(t *testing.T) {
		a := assert.New(t, false)
		f, c := newFilter(false, "")
		f.ValidMap("values", map[string]int{"first": 1, "second": 2}, validators.Great(0), localeutil.Phrase("invalid"))
		a.Empty(c)
	})

	t.Run("收集所有错误", func(t *testing.T) {
		a := assert.New(t, false)
		f, c := newFilter(false, "")
		f.ValidMap("values", map[string]int{"valid": 1, "first": 0, "second": -1}, validators.Great(0), localeutil.Phrase("invalid"))
		a.Equal(c, map[string][]localeutil.Stringer{
			"values[first]":  {localeutil.Phrase("invalid")},
			"values[second]": {localeutil.Phrase("invalid")},
		})
	})

	t.Run("遇错后停止", func(t *testing.T) {
		a := assert.New(t, false)
		f, c := newFilter(true, "")
		f.ValidMap("values", map[string]int{"invalid": 0}, validators.Great(0), localeutil.Phrase("invalid"))
		a.Equal(c, map[string][]localeutil.Stringer{
			"values[invalid]": {localeutil.Phrase("invalid")},
		})
	})
}
