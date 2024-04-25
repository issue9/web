// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v4"
)

func TestRouters(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	rs := s.Routers()
	a.NotNil(rs)

	r1 := rs.New("r1", nil)
	a.NotNil(r1)
	a.Equal(r1, rs.Get("r1")).
		Length(rs.Routers(), 1)

	rs.Remove("r1")
	a.Nil(rs.Get("r1")).
		Length(rs.Routers(), 0)
}

func TestRouters_Handle(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	rs := s.Routers()
	a.NotNil(rs)

	router := rs.New("def", nil, Recovery(http.StatusInternalServerError, s.Logs().ERROR()))
	a.NotNil(router)
	router.Get("/get1", func(ctx *Context) Responser {
		return OK("ok")
	})
	router.Get("/panic-http-error", func(ctx *Context) Responser {
		panic(NewError(http.StatusConflict, errors.New("panic")))
	})
	router.Get("/panic-error", func(ctx *Context) Responser {
		panic(errors.New("panic"))
	})
	router.Get("/panic-string", func(ctx *Context) Responser {
		panic("panic")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/get1", nil)
	router.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusOK)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/not-exists", nil)
	router.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusNotFound)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPatch, "/get1", nil)
	router.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusMethodNotAllowed)

	s.logBuf.Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/panic-http-error", nil)
	router.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusConflict).
		Contains(s.logBuf.String(), "router_test.go:44")

	s.logBuf.Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/panic-error", nil)
	router.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError).
		Contains(s.logBuf.String(), "router_test.go:47")

	s.logBuf.Reset()
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/panic-string", nil)
	router.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError).
		Contains(s.logBuf.String(), "router_test.go:50")
}
