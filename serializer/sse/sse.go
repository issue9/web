// SPDX-License-Identifier: MIT

// Package sse [SSE] 的服务端实现
//
// [SSE]: https://html.spec.whatwg.org/multipage/server-sent-events.html
package sse

import (
	"context"
	"sync"

	"github.com/issue9/web"
)

const Mimetype = "text/event-stream"

// SSE sse 事件管理
//
// T 表示用于区分不同事件源的 ID，比如按用户区分，
// 那么该类型可能是 int64 类型的用户 ID 值。
type SSE[T comparable] struct {
	status  int
	sources *sync.Map
}

// New 声明 SSE 对象
//
// status 表示正常情况下 SSE 返回的状态码。
func New[T comparable](s *web.Server, status int) *SSE[T] {
	sse := &SSE[T]{
		status:  status,
		sources: &sync.Map{},
	}
	s.Services().Add(web.StringPhrase("SSE server"), web.ServiceFunc(sse.serve))

	return sse
}

func (sse *SSE[T]) serve(ctx context.Context) error {
	<-ctx.Done()

	sse.sources.Range(func(k, v any) bool {
		v.(*Source).Close()
		sse.sources.Delete(k)
		return true
	})
	return ctx.Err()
}

// Len 当前活动的数量
func (sse *SSE[T]) Len() (size int) {
	sse.sources.Range(func(_, _ any) bool {
		size++
		return true
	})
	return size
}
