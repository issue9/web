// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package sse

import (
	"testing"

	"github.com/issue9/web/internal/bufpool"
)

func BenchmarkSource_bytes(b *testing.B) {
	s := &Source{retry: "50"}
	for range b.N {
		b := s.bytes([]string{"111", "222"}, "event", "1")
		bufpool.Put(b)
	}
}
