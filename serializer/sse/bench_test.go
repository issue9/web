// SPDX-License-Identifier: MIT

package sse

import "testing"

func BenchmarkNewMessage(b *testing.B) {
	for i := 0; i < b.N; i++ {
		m := newMessage([]string{"123", "456"}, "event", "id", 30)
		messagePool.Put(m)
	}
}

func BenchmarkMessage_bytes(b *testing.B) {
	m := &Message{Data: []string{"111", "222"}, Event: "event", ID: "1", Retry: 30}
	for i := 0; i < b.N; i++ {
		m.bytes()
	}
}
