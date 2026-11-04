// SPDX-FileCopyrightText: 2024-2026 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
	"testing"

	"github.com/issue9/assert/v6"

	"github.com/issue9/web/filter/sanitizers"
)

func TestS2R(t *testing.T) {
	a := assert.New(t, false)

	id := "x "
	v, c := newFilter(false, "")
	v.ApplyRule("id", &id, S2R(sanitizers.TrimRight), S2R(sanitizers.Upper))
	a.Length(c, 0).Equal(id, "X")
}

func TestSS2R(t *testing.T) {
	a := assert.New(t, false)

	vals := []string{"s1 ", "s2"}
	v, c := newFilter(false, "")
	v.ApplyRule("vals", &vals, SS2R[[]string](sanitizers.TrimRight, sanitizers.Upper))
	a.Empty(c).Equal(vals, []string{"S1", "S2"})
}

func TestMS2R(t *testing.T) {
	a := assert.New(t, false)

	vals := map[string]string{"s1 ": "s1 ", "s2": "s2"}
	v, c := newFilter(false, "")
	v.ApplyRule("vals", &vals, MS2R[map[string]string](sanitizers.Upper))
	a.Empty(c).Equal(vals, map[string]string{"s1 ": "S1 ", "s2": "S2"})
}
