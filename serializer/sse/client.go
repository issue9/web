// SPDX-License-Identifier: MIT

package sse

import (
	"bufio"
	"context"
	"net/http"

	"github.com/issue9/web"
)

// OnMessage 对消息的处理
func OnMessage(ctx context.Context, l web.Logger, req *http.Request, c *http.Client) (chan *Message, error) {
	if c == nil {
		c = &http.Client{}
	}

	req.Header.Set("Cache-Control", "n o-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Accept", Mimetype)

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	msg := make(chan *Message, 10)
	s := bufio.NewScanner(resp.Body)
	s.Split(bufio.ScanLines)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				m := newEmptyMessage()
				for {
					s.Scan()
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

	return msg, nil
}
