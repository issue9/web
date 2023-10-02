// SPDX-License-Identifier: MIT

package sse

import (
	"bytes"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/header"
)

const bufMaxSize = 1024

var bufPool = &sync.Pool{New: func() any { return &bytes.Buffer{} }}

type Source struct {
	data chan []byte
	exit chan struct{}
	done chan struct{}
	last time.Time
}

// Get 返回指定 ID 的事件源
//
// 仅在 [SSE.NewSource] 执行之后，此函数才能返回非空值。
func (sse *SSE[T]) Get(id T) *Source {
	if v, found := sse.sources.Load(id); found {
		return v.(*Source)
	}
	return nil
}

// NewSource 声明新的事件源
//
// NOTE: 只有采用此方法声明之后，才有可能通过 [SSE.Get] 获取实例。
// id 表示是事件源的唯一 ID，如果事件是根据用户进行区分的，那么该值应该是表示用户的 ID 值；
// wait 当前 s 退出时，wait 才会返回，可以在 [web.Handler] 中阻止路由退出。
func (sse *SSE[T]) NewSource(id T, ctx *web.Context) (s *Source, wait func()) {
	if ss, found := sse.sources.LoadAndDelete(id); found {
		ss.(*Source).Close()
	}

	s = &Source{
		data: make(chan []byte, 1),
		exit: make(chan struct{}, 1),
		done: make(chan struct{}, 1),
	}
	sse.sources.Store(id, s)

	go func() {
		s.connect(ctx, sse.status) // 阻塞，出错退出
		close(s.data)              // 退出之前关闭，防止退出之后，依然有数据源源不断地从 Sent 输入。
		sse.sources.Delete(id)     // 如果 connect 返回，说明断开了连接，删除 sources 中的记录。
	}()
	return s, s.wait
}

// 和客户端进行连接，如果返回，则表示连接被关闭。
func (s *Source) connect(ctx *web.Context, status int) {
	rc := http.NewResponseController(ctx)

	ctx.Header().Set("content-type", header.BuildContentType(Mimetype, header.UTF8Name))
	ctx.Header().Set("Content-Length", "0")
	ctx.Header().Set("Cache-Control", "no-cache")
	ctx.Header().Set("Connection", "keep-alive")
	ctx.SetCharset("utf-8")
	ctx.SetEncoding("")
	ctx.WriteHeader(status)

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
				return
			}
			if err := rc.Flush(); err != nil {
				panic(err) // 无法实现当前需要的功能，直接 panic。
			}
			s.last = time.Now()
		}
	}
}

// Sent 发送消息
//
// id、event 和  retry 都可以为空，表示不需要这些值；
func (s *Source) Sent(data []string, event, id string, retry uint) {
	w := bufPool.Get().(*bytes.Buffer)
	w.Reset()

	for _, line := range data {
		w.WriteString("data: ")
		w.WriteString(line)
		w.WriteByte('\n')
	}
	if event != "" {
		w.WriteString("event: ")
		w.WriteString(event)
		w.WriteByte('\n')
	}
	if id != "" {
		w.WriteString("id: ")
		w.WriteString(id)
		w.WriteByte('\n')
	}
	if retry > 0 {
		w.WriteString("retry: ")
		w.WriteString(strconv.Itoa(int(retry)))
		w.WriteByte('\n')
	}
	w.WriteByte('\n')

	s.data <- w.Bytes()

	if w.Cap() < bufMaxSize {
		bufPool.Put(w)
	}
}

// 关闭当前事件源
//
// 这将导致关联的 [Server.NewSource] 的 wait 直接返回。
func (s *Source) Close() { s.exit <- struct{}{} }

func (s *Source) wait() { <-s.done }
