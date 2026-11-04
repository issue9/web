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
	v, c := newFilter(false)
	v.ApplyRule("id", &id, V2R[string](localeutil.StringPhrase("required"), required))
	a.Empty(c).Equal(id, " ")

	id = "X "
	v, c = newFilter(false)
	v.ApplyRule("id", &id, S2R(sanitizers.Upper), V2R[string](localeutil.StringPhrase("required"), required))
	a.Empty(c).Equal(id, "X ")
}

func TestSV2R(t *testing.T) {
	a := assert.New(t, false)

	vals := []string{"s1 ", "s2"}
	v, c := newFilter(false)
	v.ApplyRule("vals", &vals, SV2R[[]string](localeutil.StringPhrase("required"), required))
	a.Empty(c).Equal(vals, []string{"s1 ", "s2"})

	vals = []string{"s1 ", ""}
	v, c = newFilter(false)
	v.ApplyRule("vals", &vals, SV2R[[]string](localeutil.StringPhrase("required"), required))
	a.Equal(c["vals[1]"], []localeutil.StringPhrase{localeutil.StringPhrase("required")}).
		Equal(vals, []string{"s1 ", ""})
}

func TestMV2R(t *testing.T) {
	a := assert.New(t, false)

	vals := map[string]string{"s1 ": "s1", "s2": "s2"}
	v, c := newFilter(false)
	v.ApplyRule("vals", &vals, MV2R[map[string]string](localeutil.StringPhrase("required"), required))
	a.Empty(c)

	vals = map[string]string{"s1 ": "x", "s2": ""}
	v, c = newFilter(false)
	v.ApplyRule("vals", &vals, MV2R[map[string]string](localeutil.StringPhrase("required"), required))
	a.Equal(c["vals[s2]"], []localeutil.StringPhrase{localeutil.StringPhrase("required")}).
		Equal(vals, map[string]string{"s1 ": "x", "s2": ""})
}
