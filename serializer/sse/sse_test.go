// SPDX-License-Identifier: MIT

package sse

import (
	"testing"

	"github.com/issue9/assert/v3"
)

func TestMessage_bytes(t *testing.T) {
	a := assert.New(t, false)

	a.PanicString(func() {
		m := &Message{}
		m.bytes()
	}, "data 不能为空")

	m := &Message{Data: []string{"111"}}
	a.Equal(m.bytes(), "data: 111\n\n")

	m = &Message{Data: []string{"111", "222"}}
	a.Equal(m.bytes(), "data: 111\ndata: 222\n\n")

	m = &Message{Data: []string{"111", "222"}, Event: "event"}
	a.Equal(m.bytes(), "data: 111\ndata: 222\nevent: event\n\n")

	m = &Message{Data: []string{"111", "222"}, Event: "event", ID: "1"}
	a.Equal(m.bytes(), "data: 111\ndata: 222\nevent: event\nid: 1\n\n")

	m = &Message{Data: []string{"111", "222"}, Event: "event", ID: "1", Retry: 30}
	a.Equal(m.bytes(), "data: 111\ndata: 222\nevent: event\nid: 1\nretry: 30\n\n")
}
