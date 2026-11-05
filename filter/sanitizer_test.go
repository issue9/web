// SPDX-FileCopyrightText: 2024-2026 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
	"testing"

	"github.com/issue9/assert/v6"
	"github.com/issue9/localeutil"

	"github.com/issue9/web/filter/sanitizers"
)

func TestS2R(t *testing.T) {
	a := assert.New(t, false)

	id := "x "
	v, c := newFilter(false, "")
	v.Apply("id", &id, S2R(sanitizers.TrimRight), S2R(sanitizers.Upper))
	a.Length(c, 0).Equal(id, "X")
}

func TestFilter_Sanitize(t *testing.T) {
	t.Run("正常修改", func(t *testing.T) {
		a := assert.New(t, false)
		f, c := newFilter(false, "")
		value := " x "
		f.Sanitize(&value, sanitizers.TrimRight, sanitizers.Upper)
		a.Equal(value, " X")
		a.Empty(c)
	})

	t.Run("已有错误时提前退出", func(t *testing.T) {
		a := assert.New(t, false)
		f, c := newFilter(true, "")
		f.AddReason("field", localeutil.Phrase("error"))
		value := " x "
		f.Sanitize(&value, sanitizers.Trim)
		a.Equal(value, " x ")
		a.Equal(c, map[string][]localeutil.Stringer{
			"field": {localeutil.Phrase("error")},
		})
	})
}

func TestFilter_SanitizeSlice(t *testing.T) {
	t.Run("正常修改", func(t *testing.T) {
		a := assert.New(t, false)
		f, c := newFilter(false, "")
		values := []string{" x ", "y  "}
		f.SanitizeSlice(&values, sanitizers.TrimRight, sanitizers.Upper)
		a.Equal(values, []string{" X", "Y"})
		a.Empty(c)
	})

	t.Run("已有错误时提前退出", func(t *testing.T) {
		a := assert.New(t, false)
		f, c := newFilter(true, "")
		f.AddReason("field", localeutil.Phrase("error"))
		values := []string{" x "}
		f.SanitizeSlice(&values, sanitizers.Trim)
		a.Equal(values, []string{" x "})
		a.Equal(c, map[string][]localeutil.Stringer{
			"field": {localeutil.Phrase("error")},
		})
	})
}

func TestFilter_SanitizeMap(t *testing.T) {
	t.Run("正常修改", func(t *testing.T) {
		a := assert.New(t, false)
		f, c := newFilter(false, "")
		values := map[string]string{"first": " x ", "second": "y  "}
		f.SanitizeMap(values, sanitizers.TrimRight, sanitizers.Upper)
		a.Equal(values, map[string]string{"first": " X", "second": "Y"})
		a.Empty(c)
	})

	t.Run("已有错误时提前退出", func(t *testing.T) {
		a := assert.New(t, false)
		f, c := newFilter(true, "")
		f.AddReason("field", localeutil.Phrase("error"))
		values := map[string]string{"value": " x "}
		f.SanitizeMap(values, sanitizers.Trim)
		a.Equal(values, map[string]string{"value": " x "})
		a.Equal(c, map[string][]localeutil.Stringer{
			"field": {localeutil.Phrase("error")},
		})
	})
}
