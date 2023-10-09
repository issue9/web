// SPDX-License-Identifier: MIT

package sse

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/header"
)

// Server SSE 服务端
//
// T 表示用于区分不同事件源的 ID，比如按用户区分，
// 那么该类型可能是 int64 类型的用户 ID 值。
type Server[T comparable] struct {
	retry   int64
	sources *sync.Map
}

type Source struct {
	retry int64
	data  chan []byte
	exit  chan struct{}
	done  chan struct{}
}

type Event struct {
	source  *Source
	name    string
	marshal web.MarshalFunc
}

// NewServer 声明 [Server] 对象
//
// retry 表示反馈给用户的 retry 字段，可以为零值，表示不需要输出该字段；
func NewServer[T comparable](s *web.Server, retry time.Duration) *Server[T] {
	srv := &Server[T]{
		retry:   retry.Milliseconds(),
		sources: &sync.Map{},
	}

	s.Services().Add(web.StringPhrase("SSE server"), web.ServiceFunc(srv.serve))

	return srv
}

func (srv *Server[T]) serve(ctx context.Context) error {
	<-ctx.Done()

	srv.sources.Range(func(k, v any) bool {
		v.(*Source).Close()
		srv.sources.Delete(k)
		return true
	})
	return ctx.Err()
}

// Len 当前链接的数量
func (srv *Server[T]) Len() (size int) {
	srv.sources.Range(func(_, _ any) bool {
		size++
		return true
	})
	return size
}

// Get 返回指定 ID 的事件源
//
// 仅在 [Server.NewSource] 执行之后，此函数才能返回非空值。
func (srv *Server[T]) Get(id T) *Source {
	if v, found := srv.sources.Load(id); found {
		return v.(*Source)
	}
	return nil
}

// NewSource 声明新的事件源
//
// NOTE: 只有采用此方法声明之后，才有可能通过 [Server.Get] 获取实例。
// id 表示是事件源的唯一 ID，如果事件是根据用户进行区分的，那么该值应该是表示用户的 ID 值；
// wait 当前 s 退出时，wait 才会返回，可以在 [web.Handler] 中阻止路由退出导致的 ctx 被回收。
func (srv *Server[T]) NewSource(id T, ctx *web.Context) (s *Source, wait func()) {
	if ss, found := srv.sources.LoadAndDelete(id); found {
		ss.(*Source).Close()
	}

	s = &Source{
		retry: srv.retry,
		data:  make(chan []byte, 1),
		exit:  make(chan struct{}, 1),
		done:  make(chan struct{}, 1),
	}
	srv.sources.Store(id, s)

	go func() {
		s.connect(ctx)               // 阻塞，出错退出
		defer close(s.data)          // 退出之前关闭，防止退出之后，依然有数据源源不断地从 Sent 输入。
		defer srv.sources.Delete(id) // 如果 connect 返回，说明断开了连接，删除 sources 中的记录。
	}()
	return s, s.wait
}

// 和客户端进行连接，如果返回，则表示连接被关闭。
func (s *Source) connect(ctx *web.Context) {
	ctx.Header().Set(header.ContentType, header.BuildContentType(Mimetype, header.UTF8Name))
	ctx.Header().Set(header.ContentLength, "0")
	ctx.Header().Set("Cache-Control", "no-cache")
	ctx.Header().Set("Connection", "keep-alive")
	ctx.SetCharset(header.UTF8Name)
	ctx.SetEncoding("")
	ctx.WriteHeader(http.StatusOK) // 根据标准，就是 200。

	rc := http.NewResponseController(ctx)
	for {
		select {
		case <-ctx.Request().Context().Done():
			s.done <- struct{}{}
		case <-s.exit: // 由 Source.Close 触发
			s.done <- struct{}{}
			return
		case data := <-s.data:
			if _, err := ctx.Write(data); err != nil { // 出错即退出，由客户端自行重连。
				ctx.Logs().ERROR().Error(err)
				s.done <- struct{}{}
				continue
			}

			if err := rc.Flush(); err != nil {
				if errors.Is(err, http.ErrNotSupported) {
					panic(err) // 不支持功能，直接 panic
				}
				ctx.Logs().ERROR().Error(err)
			}
		}
	}
}

// Sent 发送消息
//
// id 和 event都可以为空，表示不需要这些值；
// 如果不想输出 retry 可以输出一个非整数，按照规则客户端会忽略非整数的值；
func (s *Source) Sent(data []string, event, id string) {
	m := newMessage(data, event, id, s.retry)
	defer m.Destory()
	s.SentRaw(m.bytes())
}

// SentRaw 发送原始的数据内容
//
// NOTE: 如果有发送注释的情况，只能通过此方法。
func (s *Source) SentRaw(data []byte) { s.data <- data }

// 关闭当前事件源
//
// 这将导致关联的 [Server.NewSource] 的 wait 直接返回。
func (s *Source) Close() { s.exit <- struct{}{} }

func (s *Source) wait() { <-s.done }

// NewEvent 声明一个新的事件类型
//
// name 表示事件名称，最终输出为 event 字段；
// marshal 表示 data 字段的编码方式；
func (s *Source) NewEvent(name string, marshal web.MarshalFunc) *Event {
	return &Event{name: name, marshal: marshal, source: s}
}

func (e *Event) Sent(obj any) error {
	data, err := e.marshal(obj)
	if err != nil {
		return err
	}

	e.source.Sent(strings.Split(string(data), "\n"), e.name, "")
	return nil
}
