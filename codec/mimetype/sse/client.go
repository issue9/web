// SPDX-License-Identifier: MIT

package sse

import (
	"bufio"
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/issue9/web"
	"github.com/issue9/web/internal/header"
)

var messagePool = &sync.Pool{New: func() any { return &Message{} }}

type Message struct {
	Data  []string
	Event string
	ID    string
	Retry int64
}

func newEmptyMessage() *Message {
	m := messagePool.Get().(*Message)
	m.Data = m.Data[:0]
	m.Event = ""
	m.ID = ""
	m.Retry = 0
	return m
}

func (m *Message) append(line string) (err error) {
	prefix, data, found := strings.Cut(line, ":")
	if !found {
		return web.NewLocaleError("invalid sse format %s", line)
	}

	switch prefix {
	case "data":
		m.Data = append(m.Data, data)
	case "event":
		m.Event = data
	case "id":
		m.ID = data
	case "retry":
		m.Retry, err = strconv.ParseInt(strings.TrimSpace(data), 10, 64)
	}
	return
}

// Free 销毁当前对象
//
// NOTE: 这不是一个必须的操作，在确定不再使用当前对象的情况下，
// 执行该方法，有可能会提升一些性能。
func (m *Message) Free() { messagePool.Put(m) }

// OnMessage 对消息的处理
//
// l 用于记录运行过程的错误信息；
// msg 用于接收从服务端返回的数据对象。
// 从 msg 中取出的 [Message] 对象，在不再需要时可以调用 [Message.Free] 回收；
func OnMessage(ctx context.Context, l *web.Logger, req *http.Request, c *http.Client, msg chan *Message) error {
	if c == nil {
		c = &http.Client{}
	}

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set(header.Accept, Mimetype)

	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	s := bufio.NewScanner(resp.Body)
	s.Split(bufio.ScanLines)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				m := newEmptyMessage()
				for s.Scan() {
					if line := s.Text(); line != "" {
						if err := m.append(line); err != nil {
							l.Error(err)
						}
						continue
					}
					break // 有空行，表示已经结束一个会话。
				}
				msg <- m
			}
		}
	}()

	return nil
}
