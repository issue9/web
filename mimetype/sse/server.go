// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package sse

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/bufpool"
	"github.com/issue9/web/internal/header"
)

type (
	// Server SSE 服务端
	//
	// T 表示用于区分不同事件源的 ID，比如按用户区分，
	// 那么该类型可能是 int64 类型的用户 ID 值。
	Server[T comparable] struct {
		bufCap  int
		s       web.Server
		retry   string
		sources *sync.Map
	}

	Source struct {
		last time.Time

		lastID string
		retry  string
		buf    chan *bytes.Buffer
		exit   chan struct{}
		done   chan struct{}
	}

	SourceEvent struct {
		source  *Source
		name    string
		marshal MarshalFunc
	}

	ServerEvent[T comparable] struct {
		server  *Server[T]
		name    string
		marshal MarshalFunc
	}

	MarshalFunc = func(any) ([]byte, error)
)

// NewServer 声明 [Server] 对象
//
// retry 表示反馈给用户的 retry 字段，可以为零值，表示不需要输出该字段；
// keepAlive 表示心跳包的发送时间间隔，如果小于等于零，表示不会发送；
// bufCap 每个 SSE 队列可缓存的数据，超过此数量，调用的 Sent 将被阻塞；
func NewServer[T comparable](s web.Server, retry, keepAlive time.Duration, bufCap int) *Server[T] {
	srv := &Server[T]{
		bufCap:  bufCap,
		s:       s,
		retry:   strconv.FormatInt(retry.Milliseconds(), 10),
		sources: &sync.Map{},
	}

	s.Services().AddFunc(web.StringPhrase("SSE server"), srv.serve)
	if keepAlive > 0 {
		s.Services().AddTicker(web.StringPhrase("SSE keep alive cron"), srv.keepAlive, keepAlive, false, false)
	}

	return srv
}

func (srv *Server[T]) keepAlive(now time.Time) error {
	srv.sources.Range(func(_, v any) bool {
		if s := v.(*Source); s.last.After(now) {
			b := bufpool.New()
			b.WriteString(":\n\n")
			s.buf <- b
		}
		return true
	})
	return nil
}

func (srv *Server[T]) serve(ctx context.Context) error {
	<-ctx.Done()

	srv.sources.Range(func(_, v any) bool {
		v.(*Source).Close() // 此操作最终会从 srv.sources 中删除
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

// Get 返回指定 sid 的事件源
//
// 仅在 [Server.NewSource] 执行之后，此函数才能返回非空值。
func (srv *Server[T]) Get(sid T) *Source {
	if v, found := srv.sources.Load(sid); found {
		return v.(*Source)
	}
	return nil
}

// NewSource 声明新的事件源
//
// NOTE: 只有采用此方法声明之后，才有可能通过 [Server.Get] 获取实例。
// sid 表示是事件源的唯一 ID，如果事件是根据用户进行区分的，那么该值应该是表示用户的 ID 值；
// wait 当前 s 退出时，wait 才会返回，可以在 [web.Handler] 中阻止路由退出导致的 ctx 被回收。
func (srv *Server[T]) NewSource(sid T, ctx *web.Context) (s *Source, wait func()) {
	if ss, found := srv.sources.LoadAndDelete(sid); found {
		ss.(*Source).Close()
	}

	s = &Source{
		last: ctx.Begin(),

		lastID: ctx.Request().Header.Get("Last-Event-ID"),
		retry:  srv.retry,
		buf:    make(chan *bytes.Buffer, srv.bufCap),
		exit:   make(chan struct{}, 1),
		done:   make(chan struct{}, 1),
	}
	srv.sources.Store(sid, s)

	go func() {
		s.connect(ctx)                // 阻塞，出错退出
		defer close(s.buf)            // 退出之前关闭，防止退出之后，依然有数据源源不断地从 Sent 输入。
		defer srv.sources.Delete(sid) // 如果 connect 返回，说明断开了连接，删除 sources 中的记录。
	}()
	return s, s.wait
}

// 和客户端进行连接，如果返回，则表示连接被关闭。
func (s *Source) connect(ctx *web.Context) {
	ctx.Header().Set(header.ContentType, header.BuildContentType(Mimetype, header.UTF8Name))
	ctx.Header().Set(header.ContentLength, "0")
	ctx.Header().Set(header.CacheControl, header.NoCache)
	ctx.Header().Set(header.Connection, header.KeepAlive)
	ctx.WriteHeader(http.StatusOK) // 根据标准，就是 200。

	rc := http.NewResponseController(ctx)
	for {
		select {
		case <-ctx.Request().Context().Done():
			s.done <- struct{}{}
			return
		case <-s.exit: // 由 Source.Close 触发
			s.done <- struct{}{}
			return
		case buf := <-s.buf:
			if s.lastID != "" {
				s.lastID = ""
			}

			if _, err := ctx.Write(buf.Bytes()); err != nil {
				ctx.Logs().ERROR().Error(err)
				s.done <- struct{}{}
				continue // 出错即退出，由客户端自行重连。
			}

			if err := rc.Flush(); err != nil {
				if errors.Is(err, http.ErrNotSupported) {
					panic(err) // 不支持功能，直接 panic
				}
				ctx.Logs().ERROR().Error(err)
				s.done <- struct{}{}
				continue // 出错即退出，由客户端自行重连。
			}

			bufpool.Put(buf)
		}
	}
}

// Range 依次访问注册的 [Source] 对象
func (srv *Server[T]) Range(f func(sid T, s *Source)) {
	srv.sources.Range(func(k, v any) bool {
		f(k.(T), v.(*Source))
		return true
	})
}

// NewEvent 声明具有统一编码方式的事件派发对象
//
// name 表示事件名称，最终输出为 event 字段；
// marshal 表示 data 字段的编码方式；
func (srv *Server[T]) NewEvent(name string, marshal MarshalFunc) *ServerEvent[T] {
	return &ServerEvent[T]{
		server:  srv,
		name:    name,
		marshal: marshal,
	}
}

// Sent 向所有注册的 [Source] 发送由 f 生成的对象
func (e *ServerEvent[T]) Sent(f func(sid T, lastEventID string) any) {
	e.server.Range(func(sid T, s *Source) {
		data, err := e.marshal(f(sid, s.LastEventID()))
		if err != nil {
			e.server.s.Logs().ERROR().Error(err)
			return
		}

		s.Sent(strings.Split(string(data), "\n"), e.name, "")
	})
}

// LastEventID 客户端提交的报头 Last-Event-ID 值
//
// 当此值不为空时表示该链接刚从客户端重新连接上。
// 有新内容发送给客户端之后，该值会被重置为空。
func (s *Source) LastEventID() string { return s.lastID }

// Sent 发送消息
//
// id 和 event 都可以为空，表示不需要这些值；
// 如果不想输出 retry 可以输出一个非整数，按照规则客户端会忽略非整数的值；
func (s *Source) Sent(data []string, event, id string) { s.buf <- s.bytes(data, event, id) }

func (s *Source) bytes(data []string, event, id string) *bytes.Buffer {
	if len(data) == 0 {
		panic("data 不能为空")
	}

	w := bufpool.New()
	for _, line := range data {
		w.WriteString("data:")
		w.WriteString(line)
		w.WriteByte('\n')
	}
	if event != "" {
		w.WriteString("event:")
		w.WriteString(event)
		w.WriteByte('\n')
	}
	if id != "" {
		w.WriteString("id:")
		w.WriteString(id)
		w.WriteByte('\n')
	}

	if s.retry != "" {
		w.WriteString("retry:")
		w.WriteString(s.retry)
		w.WriteByte('\n')
	}
	w.WriteByte('\n')

	return w
}

// Close 关闭当前事件源
//
// 这将导致关联的 [Server.NewSource] 的 wait 直接返回。
func (s *Source) Close() { s.exit <- struct{}{} }

func (s *Source) wait() { <-s.done }

// NewEvent 声明具有统一编码方式的事件派发对象
//
// name 表示事件名称，最终输出为 event 字段；
// marshal 表示 data 字段的编码方式；
func (s *Source) NewEvent(name string, marshal MarshalFunc) *SourceEvent {
	return &SourceEvent{name: name, marshal: marshal, source: s}
}

func (e *SourceEvent) Sent(obj any) error {
	data, err := e.marshal(obj)
	if err != nil {
		return err
	}

	e.source.Sent(strings.Split(string(data), "\n"), e.name, "")
	return nil
}

func (e *SourceEvent) Source() *Source { return e.source }
