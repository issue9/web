// SPDX-License-Identifier: MIT

package sse

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web"
	"github.com/issue9/web/servertest"
)

var _ web.Service = &SSE[struct{}]{}

func TestEvents(t *testing.T) {
	a := assert.New(t, false)
	e := New[int64](201)
	a.NotNil(e)
	s, err := web.NewServer("test", "1.0.0", &web.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Mimetypes:  []*web.Mimetype{{Type: Mimetype}},
	})
	a.NotError(err).NotNil(s)
	s.Services().Add(web.Phrase("sse"), e)

	s.NewRouter("def", nil).Get("/events/{id}", func(ctx *web.Context) web.Responser {
		id, resp := ctx.PathInt64("id", web.ProblemBadRequest)
		if resp != nil {
			return resp
		}

		s, wait := e.NewSource(id, ctx)
		s.Sent([]string{"connect", strconv.FormatInt(id, 10)}, "", "1", 50)
		time.Sleep(time.Microsecond * 500)
		s.Sent([]string{"msg", strconv.FormatInt(id, 10)}, "event", "2", 0)
		s.Sent([]string{"msg", strconv.FormatInt(id, 10)}, "event", "2", 0)

		wait()
		return nil
	})

	defer servertest.Run(a, s)()
	defer s.Close(0)

	time.AfterFunc(5000*time.Microsecond, func() {
		e.Get(5).Close()
	})

	servertest.Get(a, "http://localhost:8080/events/5").
		Header("accept", Mimetype).
		Header("accept-encoding", "").
		Do(nil).
		Status(201).
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
