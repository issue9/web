// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"net/http"
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/internal/testdata"
	"github.com/issue9/web/serializer/json"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

func newServer(a *assert.Assertion) *Server {
	s, err := NewServer("test", "1.0.0", &Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Mimetypes: []*server.Mimetype{
			{
				Type:      "application/json",
				Marshal:   json.Marshal,
				Unmarshal: json.Unmarshal,
			},
		},
	})
	a.NotError(err).NotNil(s)

	return s
}

func TestNotModified(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a)

	r := s.NewRouter("def", nil)
	r.Any("/string", func(ctx *Context) Responser {
		const body = "string"
		return NotModified(
			func() (string, bool) { return body, true },
			func() (any, error) { return body, nil },
		)
	})
	r.Get("/bytes", func(ctx *Context) Responser {
		const body = "bytes"
		return NotModified(
			func() (string, bool) { return body, false },
			func() (any, error) { return []byte(body), nil },
		)
	})
	r.Get("/errors-500", func(ctx *Context) Responser {
		const body = "500"
		return NotModified(
			func() (string, bool) { return body, false },
			func() (any, error) { return nil, errors.New("500") },
		)
	})

	defer servertest.Run(a, s)()
	defer s.Close(0)

	// get /string

	resp := servertest.Get(a, "http://localhost:8080/string").
		Do(nil).
		Status(http.StatusOK).
		Resp()
	tag := resp.Header.Get(header.ETag)
	servertest.Get(a, "http://localhost:8080/string").
		Header(header.IfNoneMatch, tag).
		Do(nil).
		Status(http.StatusNotModified)

	// post /string

	resp = servertest.Post(a, "http://localhost:8080/string", nil).
		Do(nil).
		Status(http.StatusOK).
		Resp()
	tag = resp.Header.Get(header.ETag)
	servertest.Post(a, "http://localhost:8080/string", nil).
		Header(header.IfNoneMatch, tag).
		Do(nil).
		Status(http.StatusOK)

	// get /bytes

	resp = servertest.Get(a, "http://localhost:8080/bytes").
		Do(nil).
		Status(http.StatusOK).
		Resp()
	tag = resp.Header.Get(header.ETag)
	servertest.Get(a, "http://localhost:8080/bytes").
		Header(header.IfNoneMatch, tag).
		Do(nil).
		Status(http.StatusNotModified)

	// get /errors-500

	resp = servertest.Get(a, "http://localhost:8080/errors-500").
		Do(nil).
		Status(http.StatusInternalServerError).
		Resp()
	servertest.Get(a, "http://localhost:8080/errors-500").
		Do(nil).
		Status(http.StatusInternalServerError)
}

func TestCreated(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a)
	r := s.NewRouter("def", nil)

	// Location == ""
	r.Get("/created", func(ctx *Context) Responser {
		return Created(testdata.ObjectInst, "")
	})
	// Location == "/test"
	r.Get("/created/location", func(ctx *Context) Responser {
		return Created(testdata.ObjectInst, "/test")
	})

	defer servertest.Run(a, s)()
	defer s.Close(0)

	servertest.Get(a, "http://localhost:8080/created").Header("accept", "application/json").Do(nil).
		Status(http.StatusCreated).
		StringBody(testdata.ObjectJSONString)

	servertest.Get(a, "http://localhost:8080/created/location").Header("accept", "application/json").Do(nil).
		Status(http.StatusCreated).
		StringBody(testdata.ObjectJSONString).
		Header("Location", "/test")
}

func TestRedirect(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a)
	r := s.NewRouter("def", nil)

	r.Get("/not-implement", func(ctx *Context) Responser {
		return ctx.NotImplemented()
	})
	r.Get("/redirect", func(ctx *Context) Responser {
		return Redirect(http.StatusMovedPermanently, "https://example.com")
	})

	defer servertest.Run(a, s)()
	defer s.Close(0)

	servertest.Get(a, "http://localhost:8080/not-implement").Do(nil).Status(http.StatusNotImplemented)

	servertest.Get(a, "http://localhost:8080/redirect").Do(nil).
		Status(http.StatusOK). // http.Client.Do 会自动重定向并请求
		Header("Location", "")
}
