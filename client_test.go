// SPDX-License-Identifier: MIT

package web

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/compress"
	"github.com/issue9/web/servertest"
)

func TestClient(t *testing.T) {
	a := assert.New(t, false)

	s := newTestServer(a, nil)
	defer servertest.Run(a, s)()
	defer s.Close(500 * time.Millisecond)

	s.NewRouter("default", nil).Get("/get", func(ctx *Context) Responser {
		return OK(&object{Name: "name"})
	})

	mts := []*Mimetype{
		{Type: "application/json", Marshal: marshalJSON, Unmarshal: json.Unmarshal, ProblemType: "application/problem+json"},
	}
	cps := []*Compress{
		{Name: "gzip", Compress: compress.NewGzipCompress(3), Types: []string{"application/*"}},
	}
	c, err := NewClient("http://localhost:8080", "application/json", json.Marshal, mts, cps)
	a.NotError(err).NotNil(c)

	resp := &object{}
	p := &RFC7807{}
	a.NotError(c.Get("/get", resp, p))
	a.Zero(p).Equal(resp, &object{Name: "name"})

	resp = &object{}
	p = &RFC7807{}
	a.NotError(c.Delete("/get", resp, p))
	a.Zero(resp).Equal(p.Type, ProblemMethodNotAllowed)
}

func TestServer_NewClient(t *testing.T) {
	a := assert.New(t, false)

	s := newTestServer(a, nil)
	defer servertest.Run(a, s)()
	defer s.Close(500 * time.Millisecond)

	s.NewRouter("default", nil).Post("/post", func(ctx *Context) Responser {
		obj := &object{}
		if resp := ctx.Read(true, obj, ProblemBadRequest); resp != nil {
			return resp
		}
		return OK(obj)
	})

	s2 := newTestServer(a, nil)
	c := s2.NewClient("http://localhost:8080", "application/json", json.Marshal)
	a.NotNil(c)

	resp := &object{}
	p := &RFC7807{}
	a.NotError(c.Post("/post", &object{Age: 1}, resp, p))
	a.Zero(p).Equal(resp, &object{Age: 1})

	resp = &object{}
	p = &RFC7807{}
	a.NotError(c.Patch("/get", nil, resp, p))
	a.Zero(resp).Equal(p.Type, ProblemNotFound)
}
