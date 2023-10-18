// SPDX-License-Identifier: MIT

package web

import (
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"golang.org/x/text/encoding"

	"github.com/issue9/web/servertest"
)

// TODO
func testClient(t *testing.T) {
	a := assert.New(t, false)

	s := newTestServer(a)
	defer servertest.Run(a, s)()
	defer s.Close(500 * time.Millisecond)

	s.NewRouter("default", nil).Get("/get", func(ctx *Context) Responser {
		return OK(&object{Name: "name"})
	}).Post("/post", func(ctx *Context) Responser {
		obj := &object{}
		if resp := ctx.Read(true, obj, ProblemBadRequest); resp != nil {
			return resp
		}
		if obj.Name != "name" {
			return ctx.Problem(ProblemBadRequest).WithExtensions(&object{Name: "name"})
		}
		return OK(obj)
	})

	c := NewClient(nil, "http://localhost:8080", "application/json", json.Marshal, func(s string) (UnmarshalFunc, encoding.Encoding, error) {
		return json.Unmarshal, nil, nil
	}, func(s string, r io.Reader) (io.ReadCloser, error) {
		return nil, nil
	}, "")
	a.NotNil(c)

	resp := &object{}
	p := &RFC7807{}
	a.NotError(c.Get("/get", resp, p))
	a.Zero(p).Equal(resp, &object{Name: "name"})

	resp = &object{}
	p = &RFC7807{}
	a.NotError(c.Delete("/get", resp, p))
	a.Zero(resp).Equal(p.Type, ProblemMethodNotAllowed)

	resp = &object{}
	p = &RFC7807{Extensions: &object{}}
	a.NotError(c.Post("/post", nil, resp, p))
	a.Zero(resp).
		Equal(p.Type, ProblemBadRequest).
		Equal(p.Extensions, &object{Name: "name"})

	resp = &object{}
	p = &RFC7807{}
	a.NotError(c.Post("/post", &object{Age: 1, Name: "name"}, resp, p))
	a.Zero(p).Equal(resp, &object{Age: 1, Name: "name"})

	resp = &object{}
	p = &RFC7807{}
	a.NotError(c.Patch("/not-exists", nil, resp, p))
	a.Zero(resp).Equal(p.Type, ProblemNotFound)
}
