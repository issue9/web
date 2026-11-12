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
	"github.com/issue9/web/locales"
)

func TestFilter_Apply(t *testing.T) {
	a := assert.New(t, false)

	min2 := V2R(localeutil.Phrase("-2"), validators.GreatEqual(-2))
	min3 := V2R(localeutil.Phrase("-3"), validators.GreatEqual(-3))
	max50 := V2R(localeutil.Phrase("50"), validators.Less(50))
	max4 := V2R(localeutil.Phrase("-4"), validators.Less(-4))

	f, c := newFilter(false, "")
	n100 := -100
	p100 := 100
	f.Apply("f1", &n100, min2, min3).Apply("f2", &p100, max50, max4)
	a.Equal(c, map[string][]localeutil.Stringer{
		"f1": {localeutil.Phrase("-2")},
		"f2": {localeutil.Phrase("50")},
	})

	f, c = newFilter(true, "")
	n100 = -100
	p100 = 100
	f.Apply("f1", &n100, min2, min3).Apply("f2", &p100, max50, max4)
	a.Equal(c, map[string][]localeutil.Stringer{
		"f1": {localeutil.Phrase("-2")},
	})

	_, collector := CreateMapCollector()
	f.Reset("reset", false, collector)
	a.Equal(f.Name(), "reset").False(f.ExitOnError())
}

func TestFilter_ApplySlice(t *testing.T) {
	t.Run("正常修改", func(t *testing.T) {
		a := assert.New(t, false)
		f, c := newFilter(false, "")
		values := []int{1, -1, 3}
		f.ApplySlice("values", &values, S2R[int](sanitizers.Abs))
		a.Equal(values, []int{1, 1, 3})
		a.Empty(c)
	})

	t.Run("修改在前，之后验证出错", func(t *testing.T) {
		a := assert.New(t, false)

		f, c := newFilter(true, "")
		values := []int{1, -1, 3}
		f.ApplySlice("values", &values, S2R[int](sanitizers.Abs), V2R(locales.ShouldGreatThan(1), validators.Great(1)))
		a.Equal(c, map[string][]localeutil.Stringer{
			"values[0]": {locales.ShouldGreatThan(1)},
		})
	})

	t.Run("验证出错，之后修改", func(t *testing.T) {
		a := assert.New(t, false)

		f, c := newFilter(false, "")
		values := []int{1, -1, -3}
		f.ApplySlice("values", &values, V2R(locales.ShouldGreatThan(0), validators.Great(0)), S2R[int](sanitizers.Abs))
		a.Equal(c, map[string][]localeutil.Stringer{
			"values[1]": {locales.ShouldGreatThan(0)},
			"values[2]": {locales.ShouldGreatThan(0)},
		})
	})
}

func TestFilter_ApplyMap(t *testing.T) {
	t.Run("正常修改", func(t *testing.T) {
		a := assert.New(t, false)
		f, c := newFilter(false, "")
		values := map[string]int{"first": 1, "second": -1, "third": 3}
		f.ApplyMap("values", values, S2R[int](sanitizers.Abs))
		a.Equal(values, map[string]int{"first": 1, "second": 1, "third": 3})
		a.Empty(c)
	})

	t.Run("修改在前，之后验证出错", func(t *testing.T) {
		a := assert.New(t, false)
		f, c := newFilter(true, "")
		values := map[string]int{"first": 1, "second": -2, "third": 3}
		f.ApplyMap("values", values, S2R[int](sanitizers.Abs), V2R(locales.ShouldGreatThan(1), validators.Great(1)))
		a.Equal(c, map[string][]localeutil.Stringer{
			"values[first]": {locales.ShouldGreatThan(1)},
		})
	})

	t.Run("验证出错，之后修改", func(t *testing.T) {
		a := assert.New(t, false)
		f, c := newFilter(true, "")
		values := map[string]int{"first": 1, "second": 0, "third": 3}
		f.ApplyMap("values", values, V2R(locales.ShouldGreatThan(0), validators.Great(0)), S2R[int](sanitizers.Abs))
		a.Equal(c, map[string][]localeutil.Stringer{
			"values[second]": {locales.ShouldGreatThan(0)},
		})
	})
}
