// SPDX-License-Identifier: MIT

package sse

import "testing"

func BenchmarkSource_bytes(b *testing.B) {
	s := &Source{retry: "50"}
	for i := 0; i < b.N; i++ {
		b := s.bytes([]string{"111", "222"}, "event", "1")
		bufPool.Put(b)
	}
}
