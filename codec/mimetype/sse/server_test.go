// SPDX-License-Identifier: MIT

package sse

import (
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web"
	"github.com/issue9/web/codec"
	"github.com/issue9/web/codec/mimetype/json"
	"github.com/issue9/web/logs"
	"github.com/issue9/web/server"
	"github.com/issue9/web/servertest"
)

func TestServer(t *testing.T) {
	a := assert.New(t, false)
	s, err := server.New("test", "1.0.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Mimetypes: []*codec.Mimetype{
			{Name: "application/json", MarshalBuilder: json.BuildMarshal, Unmarshal: json.Unmarshal},
		},
		Logs: &logs.Options{
			Created: logs.MicroLayout,
			Handler: logs.NewTermHandler(os.Stderr, nil),
			Levels:  logs.AllLevels(),
		},
	})
	a.NotError(err).NotNil(s)
	e := NewServer[int64](s, 50*time.Millisecond, 5*time.Second, 10)
	a.NotNil(e)
	s.NewRouter("default", nil).Get("/event/{id}", func(ctx *web.Context) web.Responser {
		id, resp := ctx.PathInt64("id", web.ProblemBadRequest)
		if resp != nil {
			return resp
		}

		a.Equal(0, e.Len())

		s, wait := e.NewSource(id, ctx)
		s.Sent([]string{"connect", strconv.FormatInt(id, 10)}, "", "1")
		time.Sleep(time.Microsecond * 500)

		event := s.NewEvent("event", json.BuildMarshal(nil))
		event.Sent(1)
		time.Sleep(time.Microsecond * 500)
		event.Sent(&struct{ ID int }{ID: 5})

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
		Header("accept", "application/json").
		Header("accept-encoding", "").
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

	a.Equal(s.bytes([]string{"111"}, "", "").String(), "data:111\n\n")

	a.Equal(s.bytes([]string{"111", "222"}, "", "").String(), "data:111\ndata:222\n\n")

	a.Equal(s.bytes([]string{"111", "222"}, "event", "").String(), "data:111\ndata:222\nevent:event\n\n")

	a.Equal(s.bytes([]string{"111", "222"}, "event", "1").String(), "data:111\ndata:222\nevent:event\nid:1\n\n")

	s.retry = "30"
	a.Equal(s.bytes([]string{"111", " 222"}, "event", " 1").String(), "data:111\ndata: 222\nevent:event\nid: 1\nretry:30\n\n")
}
