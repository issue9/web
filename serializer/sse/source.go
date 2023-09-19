// SPDX-License-Identifier: MIT

package sse

import (
	"net/http"
	"strconv"

	"github.com/issue9/errwrap"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/header"
)

type Source struct {
	data chan []byte
	exit chan struct{}
	done chan struct{}
}

// Get 返回指定 ID 的事件源
func (sse *SSE[T]) Get(id T) *Source { return sse.sources[id] }

// NewSource 声明新的事件源
//
// NOTE: 只有采用此方法声明之后，才有可能通过 [Server.Get] 获取实例。
// id 表示是事件源的唯一 ID，如果事件是根据用户进行区分的，那么该值应该是表示用户的 ID 值；
// wait 当前 s 退出时，wait 才会返回，可以在 [web.Handler] 中阻止路由退出。
func (sse *SSE[T]) NewSource(id T, ctx *web.Context) (s *Source, wait func()) {
	if sse.sources[id] != nil {
		sse.sources[id].Close()
	}

	s = &Source{
		data: make(chan []byte, 1),
		exit: make(chan struct{}, 1),
		done: make(chan struct{}, 1),
	}
	sse.sources[id] = s

	go func() {
		s.connect(ctx, sse.status) // 阻塞，出错退出
		close(s.data)              // 退出之前关闭，防止退出之后，依然有数据源源不断地从 Sent 输入。
		delete(sse.sources, id)    // 如果 connect 返回，说明断开了连接，删除 sources 中的记录。
	}()
	return s, s.wait
}

// 和客户端进行连接，如果返回，则表示连接被关闭。
func (s *Source) connect(ctx *web.Context, status int) {
	var rw http.ResponseWriter = ctx
	f, ok := rw.(http.Flusher)
	for !ok { // TODO: go1.20 之后，可以采用 http.ResponseController 方法。
		if rr, rok := rw.(interface{ Unwrap() http.ResponseWriter }); rok {
			rw = rr.Unwrap()
			f, ok = rw.(http.Flusher)
			continue
		}
		break
	}
	if f == nil {
		panic("ctx 无法转换成 http.Flusher") // 无法实现当前需要的功能，直接 panic。
	}

	ctx.Header().Set("content-type", header.BuildContentType(Mimetype, header.UTF8Name))
	ctx.Header().Set("Content-Length", "0")
	ctx.Header().Set("Cache-Control", "no-cache")
	ctx.Header().Set("Connection", "keep-alive")
	ctx.SetCharset("utf-8")
	ctx.SetEncoding("")
	ctx.WriteHeader(status)

	for {
		select {
		case <-s.exit:
			s.done <- struct{}{}
			return
		case data := <-s.data:
			if _, err := ctx.Write(data); err != nil { // 出错即退出，由客户端自行重连。
				ctx.Logs().ERROR().Error(err)
				s.done <- struct{}{}
				return
			}
			f.Flush()
		}
	}
}

// Sent 发送消息
//
// id、event 和  retry 都可以为空，表示不需要这些值；
func (s *Source) Sent(d []string, event, id string, retry uint) {
	w := errwrap.Buffer{}
	for _, line := range d {
		w.WString("data: ").WString(line).WByte('\n')
	}
	if event != "" {
		w.WString("event: ").WString(event).WByte('\n')
	}
	if id != "" {
		w.WString("id: ").WString(id).WByte('\n')
	}
	if retry > 0 {
		w.WString("retry: ").WString(strconv.Itoa(int(retry))).WByte('\n')
	}
	w.WByte('\n')

	s.data <- w.Bytes()
}

// 关闭当前事件源
//
// 这将导致关联的 [Server.NewSource] 的 wait 直接返回。
func (s *Source) Close() { s.exit <- struct{}{} }

func (s *Source) wait() { <-s.done }
