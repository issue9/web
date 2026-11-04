// SPDX-FileCopyrightText: 2024-2026 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
	"errors"
	"testing"

	"github.com/issue9/assert/v6"
	"github.com/issue9/localeutil"

	"github.com/issue9/web/filter/sanitizers"
)

func TestRules(t *testing.T) {
	a := assert.New(t, false)

	id := "x "
	v, c := newFilter(false, "")
	v.ApplyRule("id", &id, Rules(S2R(sanitizers.TrimRight), S2R(sanitizers.Upper)))
	a.Length(c, 0).Equal(id, "X")
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
