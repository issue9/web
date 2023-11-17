// SPDX-License-Identifier: MIT

package sse

import (
	"context"
	sj "encoding/json"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web"
	"github.com/issue9/web/codec/mimetype/json"
	"github.com/issue9/web/logs"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

func TestMessage_append(t *testing.T) {
	a := assert.New(t, false)

	m := newEmptyMessage()
	a.NotError(m.append("data:abc")).
		Equal(m.Data, []string{"abc"})

	m = newEmptyMessage()
	a.NotError(m.append("data:abc")).
		NotError(m.append("data:abc")).
		Equal(m.Data, []string{"abc", "abc"})

	m = newEmptyMessage()
	a.NotError(m.append("event:abc")).
		Equal(m.Event, "abc")

	m = newEmptyMessage()
	a.NotError(m.append("event:abc")).
		NotError(m.append("id: 123")).
		Equal(m.Event, "abc").
		Equal(m.ID, " 123")

	m = newEmptyMessage()
	a.NotError(m.append("event:abc")).
		NotError(m.append("id: 123")).
		NotError(m.append("retry: 123")).
		Equal(m.Event, "abc").
		Equal(m.ID, " 123").
		Equal(m.Retry, 123)
}

func TestOnMessage(t *testing.T) {
	a := assert.New(t, false)
	s, err := server.New("test", "1.0.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Mimetypes: []*server.Mimetype{
			{Name: "application/json", Marshal: json.Marshal, Unmarshal: json.Unmarshal},
			{Name: Mimetype, Marshal: nil, Unmarshal: nil},
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

		s, wait := e.NewSource(id, ctx)
		s.Sent([]string{"connect", strconv.FormatInt(id, 10)}, "", "1")
		time.Sleep(time.Microsecond * 500)

		event := s.NewEvent("event", sj.Marshal)
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
	msg5 := make(chan *Message, 10)
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/event/5", nil)
	a.NotError(err).NotNil(req)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = OnMessage(ctx, s.Logs().ERROR(), req, nil, msg5)
	a.NotError(err).Equal(1, e.Len())

	a.Equal(<-msg5, &Message{Data: []string{"connect", "5"}, ID: "1", Retry: 50})
	a.Equal(<-msg5, &Message{Data: []string{"1"}, Event: "event", Retry: 50})
	a.Equal(<-msg5, &Message{Data: []string{"{\"ID\":5}"}, Event: "event", Retry: 50})

	// get /event/6

	msg6 := make(chan *Message, 10)
	req, err = http.NewRequest(http.MethodGet, "http://localhost:8080/event/6", nil)
	a.NotError(err).NotNil(req)
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	err = OnMessage(ctx, s.Logs().ERROR(), req, nil, msg6)
	a.NotError(err).Equal(2, e.Len())

	a.Equal(<-msg6, &Message{Data: []string{"connect", "6"}, ID: "1", Retry: 50})
	a.Equal(<-msg6, &Message{Data: []string{"1"}, Event: "event", Retry: 50})
	a.Equal(<-msg6, &Message{Data: []string{"{\"ID\":5}"}, Event: "event", Retry: 50})

	// server event

	event := e.NewEvent("se", sj.Marshal)
	event.Sent(func(sid int64, lastID string) any {
		return &struct {
			ID     int64
			LastID string
		}{ID: sid, LastID: lastID}
	})

	a.Equal(<-msg5, &Message{Data: []string{"{\"ID\":5,\"LastID\":\"\"}"}, Event: "se", Retry: 50})
	a.Equal(<-msg6, &Message{Data: []string{"{\"ID\":6,\"LastID\":\"\"}"}, Event: "se", Retry: 50})
}
