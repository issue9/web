// SPDX-FileCopyrightText: 2024-2026 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
	"testing"

	"github.com/issue9/assert/v6"

	"github.com/issue9/web/filter/sanitizers"
)

func TestRules(t *testing.T) {
	a := assert.New(t, false)

	id := "x "
	v, c := newFilter(false, "")
	v.ApplyRule("id", &id, Rules(S2R(sanitizers.TrimRight), S2R(sanitizers.Upper)))
	a.Length(c, 0).Equal(id, "X")
}
