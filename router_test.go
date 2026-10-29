// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"net/http"
	"slices"
	"testing"

	"github.com/issue9/assert/v6"
	"github.com/issue9/mux/v10/header"
)

func TestRouters(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a, false)
	rs := s.Routers()
	a.NotNil(rs)

	r1 := rs.New("r1", nil)
	a.NotNil(r1)
	a.Equal(r1, rs.Get("r1")).
		Length(slices.Collect(rs.Routers()), 1)

	rs.Remove("r1")
	a.Nil(rs.Get("r1")).
		Length(slices.Collect(rs.Routers()), 0)
}

func TestRouters_Handle(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a, false)
	rs := s.Routers()
	a.NotNil(rs)

	router := rs.New("def", nil, withRecovery(s.Server))
	a.NotNil(router)
	router.Get("/get1", func(ctx *Context) Responser {
		return OK("ok")
	})
	router.Get("/panic-new-problem-error", func(ctx *Context) Responser {
		panic(NewProblemError(ProblemServiceUnavailable, errors.New("ProblemServiceUnavailable")))
	})
	router.Get("/panic-error", func(ctx *Context) Responser {
		panic(errors.New("panic"))
	})
	router.Get("/panic-string", func(ctx *Context) Responser {
		panic("panic")
	})

	defer s.Run()()
	defer s.Close(0)

	s.Get("/get1").Do().Success()
	s.Get("/not-exists").Do().Status(http.StatusNotFound)
	s.Patch("/get1").Do().Status(http.StatusMethodNotAllowed)

	s.logBuf.Reset()
	s.Get("/panic-new-problem-error").Header(header.AcceptEncoding, "").Do().
		Status(http.StatusServiceUnavailable)
	a.Contains(s.logBuf.String(), "router_test.go:45")

	s.logBuf.Reset()
	s.Get("/panic-error").Header(header.AcceptEncoding, "").Do().
		Status(http.StatusInternalServerError)
	a.Contains(s.logBuf.String(), "router_test.go:48")

	s.logBuf.Reset()
	s.Get("/panic-string").Header(header.AcceptEncoding, "").Do().
		Status(http.StatusInternalServerError)
	a.Contains(s.logBuf.String(), "router_test.go:51")
}
