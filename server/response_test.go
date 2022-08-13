// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/assert/v3/rest"
	"github.com/issue9/logs/v4"
)

func TestStatus(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, nil)

	a.PanicString(func() {
		Status(http.StatusOK, "k", "v", "K2")
	}, "kv 必须偶数位")

	resp := Status(http.StatusAccepted)
	a.NotNil(resp)

	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	ctx := s.newContext(w, r, nil)
	resp.Apply(ctx)
	a.Equal(w.Result().StatusCode, http.StatusAccepted)

	resp = Status(http.StatusAccepted, "k1", "v1", "k2", "v2")
	a.NotNil(resp)
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Request()
	ctx = s.newContext(w, r, nil)
	resp.Apply(ctx)
	a.Equal(w.Result().StatusCode, http.StatusAccepted).
		Equal(ctx.Header().Get("k1"), "v1").
		Equal(ctx.Header().Get("k2"), "v2")
}

func TestObject(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a, nil)

	a.PanicString(func() {
		Object(http.StatusOK, nil, "k", "v", "K2")
	}, "kv 必须偶数位")

	resp := Object(http.StatusAccepted, "obj")
	a.NotNil(resp)
	w := httptest.NewRecorder()
	r := rest.Get(a, "/path").Request()
	ctx := s.newContext(w, r, nil)
	resp.Apply(ctx)
	a.Equal(w.Result().StatusCode, http.StatusAccepted).
		Equal(w.Body.String(), `"obj"`)

	resp = Object(http.StatusAccepted, struct{ K1 string }{K1: "V1"}, "k1", "v1", "k2", "v2")
	a.NotNil(resp)
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Request()
	ctx = s.newContext(w, r, nil)
	resp.Apply(ctx)
	a.Equal(w.Result().StatusCode, http.StatusAccepted).
		Equal(ctx.Header().Get("k1"), "v1").
		Equal(ctx.Header().Get("k2"), "v2").
		Equal(w.Body.String(), `{"K1":"V1"}`)

	resp = Object(http.StatusAccepted, make(chan int))
	a.NotNil(resp)
	w = httptest.NewRecorder()
	r = rest.Get(a, "/path").Request()
	ctx = s.newContext(w, r, nil)
	resp.Apply(ctx)
	a.Empty(w.Body.String())
}

func TestContext_Error(t *testing.T) {
	a := assert.New(t, false)
	errLog := new(bytes.Buffer)

	srv := newServer(a, &Options{
		Logs: logs.New(logs.NewTextWriter("20060102-15:04:05", errLog), logs.Created, logs.Caller),
	})
	errLog.Reset()

	t.Run("Error", func(t *testing.T) {
		a := assert.New(t, false)
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		ctx := srv.newContext(w, r, nil)
		ctx.Error(http.StatusNotImplemented, errors.New("log1 log2")).Apply(ctx)
		a.Contains(errLog.String(), "response_test.go:96") // NOTE: 此测试依赖上一行的行号
		a.Contains(errLog.String(), "log1 log2")
		a.Equal(w.Code, http.StatusNotImplemented)
	})

	t.Run("InternalServerError", func(t *testing.T) {
		a := assert.New(t, false)
		errLog.Reset()
		w := httptest.NewRecorder()
		r := rest.Get(a, "/path").Request()
		ctx := srv.newContext(w, r, nil)
		ctx.InternalServerError(errors.New("log1 log2")).Apply(ctx)
		a.Contains(errLog.String(), "response_test.go:108") // NOTE: 此测试依赖上一行的行号
		a.Contains(errLog.String(), "log1 log2")
		a.Equal(w.Code, http.StatusInternalServerError)
	})
}
