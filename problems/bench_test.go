// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package problems

import (
	"slices"
	"testing"

	"github.com/issue9/assert/v6"
)

func BenchmarkGet(b *testing.B) {
	a := assert.New(b, false)
	ps := New(builder)

	var i int
	for b.Loop() {
		i = slices.IndexFunc(ps.problems, func(item *Item) bool { return item.ID == BadGateway })
	}
	a.True(i > -1)
}
