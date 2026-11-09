// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package problems

import (
	"testing"

	"github.com/issue9/assert/v6"
)

func BenchmarkProblems_Get(b *testing.B) {
	a := assert.New(b, false)
	ps := New()

	var item *Item
	for b.Loop() {
		item = ps.Get(BadGateway)
	}
	a.NotNil(item)
}
