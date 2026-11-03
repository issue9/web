// SPDX-FileCopyrightText: 2024-2026 caixw
//
// SPDX-License-Identifier: MIT

package filter

import (
	"testing"

	"github.com/issue9/assert/v6"
)

func TestRules(t *testing.T) {
	a := assert.New(t, false)

	id := "x "
	v, c := newFilter(false)
	v.ApplyRule("id", &id, Rules(S2R(trimRight), S2R(upper)))
	a.Length(c, 0).Equal(id, "X")
}
