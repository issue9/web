// SPDX-License-Identifier: MIT

// Package sse server sent event 的实现
package sse

import "context"

const Mimetype = "text/event-stream"

// SSE sse 事件管理
//
// T 表示用于区分不同事件源的 ID，比如按用户区分，那么该类型可能是 int64 类型的用户 ID 值。
type SSE[T comparable] struct {
	status  int
	sources map[T]*Source
}

func New[T comparable](status int) *SSE[T] {
	return &SSE[T]{status: status}
}

func (sse *SSE[T]) Serve(ctx context.Context) error {
	sse.sources = make(map[T]*Source, 10)

	<-ctx.Done()
	for _, s := range sse.sources {
		s.Close()
	}
	sse.sources = nil
	return ctx.Err()
}

// Len 当前活动的数量
func (sse *SSE[T]) Len() int { return len(sse.sources) }
