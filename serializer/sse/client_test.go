// SPDX-License-Identifier: MIT

package sse

import (
	"context"
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

func TestOnMessage(t *testing.T) {
	a := assert.New(t, false)
	s, err := web.NewServer("test", "1.0.0", &web.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Mimetypes: []*web.Mimetype{
			{Type: "application/json", MarshalBuilder: json.BuildMarshal, Unmarshal: json.Unmarshal},
			{Type: Mimetype, MarshalBuilder: nil, Unmarshal: nil},
		},
		Logs: &logs.Options{
			Handler: logs.NewTermHandler(logs.MicroLayout, os.Stderr, nil),
			Caller:  true,
			Created: true,
			Levels:  logs.AllLevels(),
		},
	})
	a.NotError(err).NotNil(s)
	e := NewServer[int64](s, 50*time.Millisecond)
	a.NotNil(e)
	s.NewRouter("default", nil).Get("/event/{id}", func(ctx *web.Context) web.Responser {
		id, resp := ctx.PathInt64("id", web.ProblemBadRequest)
		if resp != nil {
			return resp
		}

		s, wait := e.NewSource(id, ctx)
		s.Sent([]string{"connect", strconv.FormatInt(id, 10)}, "", "1")
		time.Sleep(time.Microsecond * 500)

		event := s.NewEvent("event", json.BuildMarshal(nil))
		event.Sent(1)
		time.Sleep(time.Microsecond * 500)
		event.Sent(&struct{ ID int }{ID: 5})

		wait()
		return nil
	})

	defer servertest.Run(a, s)()
	defer s.Close(500 * time.Millisecond)

	// get /event/5

	a.Equal(0, e.Len())
	msg := make(chan *Message, 10)
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/event/5", nil)
	a.NotError(err).NotNil(req)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = OnMessage(ctx, s.Logs().ERROR(), req, nil, msg)
	a.NotError(err).Equal(1, e.Len())

	a.Equal(<-msg, &Message{Data: []string{"connect", "5"}, ID: "1", Retry: 50})
	a.Equal(<-msg, &Message{Data: []string{"1"}, Event: "event", Retry: 50})
	a.Equal(<-msg, &Message{Data: []string{"{\"ID\":5}"}, Event: "event", Retry: 50})

	// get /event/6

	msg = make(chan *Message, 10)
	req, err = http.NewRequest(http.MethodGet, "http://localhost:8080/event/6", nil)
	a.NotError(err).NotNil(req)
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	err = OnMessage(ctx, s.Logs().ERROR(), req, nil, msg)
	a.NotError(err).Equal(2, e.Len())

	a.Equal(<-msg, &Message{Data: []string{"connect", "6"}, ID: "1", Retry: 50})
	a.Equal(<-msg, &Message{Data: []string{"1"}, Event: "event", Retry: 50})
	a.Equal(<-msg, &Message{Data: []string{"{\"ID\":5}"}, Event: "event", Retry: 50})
}
