// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package sse

import (
	sj "encoding/json"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
	"github.com/issue9/mux/v8/header"

	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/json"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

func TestServer(t *testing.T) {
	a := assert.New(t, false)
	s, err := server.New("test", "1.0.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Mimetypes: []*server.Mimetype{
			{Name: header.JSON, Marshal: json.Marshal, Unmarshal: json.Unmarshal},
		},
		Logs: &server.Logs{
			Created: server.MicroLayout,
			Handler: server.NewTermHandler(os.Stderr, nil),
			Levels:  server.AllLevels(),
		},
	})
	a.NotError(err).NotNil(s)
	e := NewServer[int64](s, 50*time.Millisecond, 5*time.Second, 10)
	a.NotNil(e)
	s.Routers().New("default", nil).Get("/event/{id}", func(ctx *web.Context) web.Responser {
		id, resp := ctx.PathInt64("id", web.ProblemBadRequest)
		if resp != nil {
			return resp
		}

		a.Equal(0, e.Len())

		s, wait := e.NewSource(id, ctx)
		s.Sent([]string{"connect", strconv.FormatInt(id, 10)}, "", "1")
		time.Sleep(time.Microsecond * 500)

		event := s.NewEvent("event", sj.Marshal)
		a.NotError(event.Sent(1))
		time.Sleep(time.Microsecond * 500)
		a.NotError(event.Sent(&struct{ ID int }{ID: 5}))

		a.Equal(1, e.Len())
		wait()
		return nil
	})

	defer servertest.Run(a, s)()
	a.Equal(0, e.Len())

	time.AfterFunc(500*time.Millisecond, func() {
		a.Nil(e.Get(100))
		a.NotNil(e.Get(5)) // 需要有人访问 /events/{id} 之后才有值
		s.Close(500 * time.Millisecond)
	})

	servertest.Get(a, "http://localhost:8080/event/5").
		Header(header.Accept, header.JSON).
		Header(header.AcceptEncoding, "").
		Do(nil).
		Status(http.StatusOK).
		StringBody(`data:connect
data:5
id:1
retry:50

data:1
event:event
retry:50

data:{"ID":5}
event:event
retry:50

`)
}

func TestSource_bytes(t *testing.T) {
	a := assert.New(t, false)
	s := &Source{}

	a.PanicString(func() {
		s.bytes(nil, "", "")
	}, "data 不能为空")

	a.Equal(s.bytes([]string{"111"}, "", "").String(), "data:111\n\n").
		Equal(s.bytes([]string{"111", "222"}, "", "").String(), "data:111\ndata:222\n\n").
		Equal(s.bytes([]string{"111", "222"}, "event", "").String(), "data:111\ndata:222\nevent:event\n\n").
		Equal(s.bytes([]string{"111", "222"}, "event", "1").String(), "data:111\ndata:222\nevent:event\nid:1\n\n")

	s.retry = "30"
	a.Equal(s.bytes([]string{"111", " 222"}, "event", " 1").String(), "data:111\ndata: 222\nevent:event\nid: 1\nretry:30\n\n")
}
