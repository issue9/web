// SPDX-License-Identifier: MIT

// Package sse [SSE] 的实现
//
// [SSE]: https://html.spec.whatwg.org/multipage/server-sent-events.html
package sse

import (
	"bytes"
	"strconv"
	"strings"
	"sync"

	"github.com/issue9/web"
)

const Mimetype = "text/event-stream"

const bufMaxSize = 1024

var (
	bufPool     = &sync.Pool{New: func() any { return &bytes.Buffer{} }}
	messagePool = &sync.Pool{New: func() any { return &Message{} }}
)

type Message struct {
	Data  []string
	Event string
	ID    string
	Retry int64
}

func newMessage(data []string, event, id string, retry int64) *Message {
	m := messagePool.Get().(*Message)
	m.Data = data
	m.Event = event
	m.ID = id
	m.Retry = retry
	return m
}

func (m *Message) bytes() []byte {
	if len(m.Data) == 0 {
		panic("data 不能为空")
	}

	w := bufPool.Get().(*bytes.Buffer)
	w.Reset()
	defer func() {
		if w.Cap() < bufMaxSize {
			bufPool.Put(w)
		}
	}()

	for _, line := range m.Data {
		w.WriteString("data: ")
		w.WriteString(line)
		w.WriteByte('\n')
	}
	if m.Event != "" {
		w.WriteString("event: ")
		w.WriteString(m.Event)
		w.WriteByte('\n')
	}
	if m.ID != "" {
		w.WriteString("id: ")
		w.WriteString(m.ID)
		w.WriteByte('\n')
	}

	if m.Retry > 0 {
		w.WriteString("retry: ")
		w.WriteString(strconv.FormatInt(m.Retry, 10))
		w.WriteByte('\n')
	}
	w.WriteByte('\n')

	return w.Bytes()
}

func (m *Message) unmarshal(u web.UnmarshalFunc, v any) error {
	data := []byte(strings.Join(m.Data, "\n"))
	return u(data, v)
}
