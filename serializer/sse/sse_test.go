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
	"github.com/issue9/web/logs"
	"github.com/issue9/web/serializer/json"
	"github.com/issue9/web/servertest"
)

func TestSSE(t *testing.T) {
	a := assert.New(t, false)
	s, err := web.NewServer("test", "1.0.0", &web.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Mimetypes: []*web.Mimetype{
			{Type: "application/json", Marshal: json.Marshal, Unmarshal: json.Unmarshal},
		},
		Logs: &logs.Options{
			Handler: logs.NewTermHandler(logs.MicroLayout, os.Stderr, nil),
			Caller:  true,
			Created: true,
			Levels:  logs.AllLevels(),
		},
	})
	a.NotError(err).NotNil(s)
	e := New[int64](s, http.StatusCreated, time.Minute, time.Minute)
	a.NotNil(e)
	s.NewRouter("default", nil).Get("/event/{id}", func(ctx *web.Context) web.Responser {
		id, resp := ctx.PathInt64("id", web.ProblemBadRequest)
		if resp != nil {
			return resp
		}

		a.Equal(0, e.Len())

		s, wait := e.NewSource(id, ctx)
		s.Sent([]string{"connect", strconv.FormatInt(id, 10)}, "", "1", 50)
		time.Sleep(time.Microsecond * 500)
		s.Sent([]string{"msg", strconv.FormatInt(id, 10)}, "event", "2", 0)
		s.Sent([]string{"msg", strconv.FormatInt(id, 10)}, "event", "2", 0)

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
		Status(http.StatusCreated).
		StringBody(`data: connect
data: 5
id: 1
retry: 50

data: msg
data: 5
event: event
id: 2

data: msg
data: 5
event: event
id: 2

`)
}
