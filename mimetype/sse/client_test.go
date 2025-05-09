// SPDX-FileCopyrightText: 2018-2025 caixw
//
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

	"github.com/issue9/assert/v4"
	"github.com/issue9/logs/v7"

	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/nop"
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
	s, err := server.NewHTTP("test", "1.0.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Codec:      web.NewCodec().AddMimetype(Mimetype, nop.Marshal, nop.Unmarshal, "", true, true),
		Logs:       logs.New(logs.NewTermHandler(os.Stderr, nil), logs.WithCreated(logs.MicroLayout)),
	})
	a.NotError(err).NotNil(s)
	e := NewServer[int64](s, 50*time.Millisecond, 5*time.Second, 10, web.StringPhrase("sse"))
	a.NotNil(e)
	s.Routers().New("default", nil).Get("/event/{id}", func(ctx *web.Context) web.Responser {
		id, resp := ctx.PathInt64("id", web.ProblemBadRequest)
		if resp != nil {
			return resp
		}

		s, wait := e.NewSource(id, ctx)
		s.Sent([]string{"connect", strconv.FormatInt(id, 10)}, "", "1")
		time.Sleep(time.Microsecond * 500)

		event := s.NewEvent("event", sj.Marshal)
		a.NotError(event.Sent(1))
		time.Sleep(time.Microsecond * 500)
		a.NotError(event.Sent(&struct{ ID int }{ID: 5}))

		wait()
		return nil
	})

	defer servertest.Run(a, s)()
	defer s.Close(500 * time.Millisecond)

	// get /event/5

	a.Equal(0, e.Len())
	msg5 := make(chan *Message, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = OnMessage(ctx, s.Logs().ERROR(), "http://localhost:8080/event/5", nil, msg5)
	a.NotError(err).Equal(1, e.Len())

	a.Equal(<-msg5, &Message{Data: []string{"connect", "5"}, ID: "1", Retry: 50}).
		Equal(<-msg5, &Message{Data: []string{"1"}, Event: "event", Retry: 50}).
		Equal(<-msg5, &Message{Data: []string{"{\"ID\":5}"}, Event: "event", Retry: 50})

	// get /event/6

	msg6 := make(chan *Message, 10)
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	err = OnMessage(ctx, s.Logs().ERROR(), "http://localhost:8080/event/6", nil, msg6)
	a.NotError(err).Equal(2, e.Len())

	a.Equal(<-msg6, &Message{Data: []string{"connect", "6"}, ID: "1", Retry: 50}).
		Equal(<-msg6, &Message{Data: []string{"1"}, Event: "event", Retry: 50}).
		Equal(<-msg6, &Message{Data: []string{"{\"ID\":5}"}, Event: "event", Retry: 50})

	// server event

	event := e.NewEvent("se", sj.Marshal)
	event.Sent(func(sid int64, lastID string) any {
		return &struct {
			ID     int64
			LastID string
		}{ID: sid, LastID: lastID}
	})

	a.Equal(<-msg5, &Message{Data: []string{"{\"ID\":5,\"LastID\":\"\"}"}, Event: "se", Retry: 50}).
		Equal(<-msg6, &Message{Data: []string{"{\"ID\":6,\"LastID\":\"\"}"}, Event: "se", Retry: 50})
}
