// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"net/http"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
)

func TestMiddlewares(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)
	app.errorhandlers.Add(func(w http.ResponseWriter, status int) {
		w.WriteHeader(status)
		w.Write([]byte("error handler test"))
	}, http.StatusNotFound)

	m1 := app.NewModule("m1", "m1 desc", "m2")
	a.NotNil(m1)
	m2 := app.NewModule("m2", "m2 desc")
	a.NotNil(m2)
	m1.GetFunc("/m1/test", f202)
	m2.GetFunc("/m2/test", f202)
	app.Mux().GetFunc("/mux/test", f202)

	a.NotError(app.Init("", nil))
	go func() {
		err := app.Serve()
		a.ErrorType(err, http.ErrServerClosed, "assert.ErrorType 错误，%v", err.Error())
	}()
	time.Sleep(500 * time.Microsecond)

	// static 中定义的静态文件
	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8082/client/file1.txt").
		Header("Accept-Encoding", "gzip,deflate").
		Do().
		Status(http.StatusOK).
		Header("Content-Type", "text/plain; charset=utf-8").
		Header("Content-Encoding", "gzip")
}

func TestDebug(t *testing.T) {
	srv := rest.NewServer(t, debug(http.HandlerFunc(f202)), nil)

	// 命中 /debug/pprof/cmdline
	srv.NewRequest(http.MethodGet, "/debug/pprof/").
		Do().
		Status(http.StatusOK)

	srv.NewRequest(http.MethodGet, "/debug/pprof/cmdline").
		Do().
		Status(http.StatusOK)

	srv.NewRequest(http.MethodGet, "/debug/pprof/trace").
		Do().
		Status(http.StatusOK)

	srv.NewRequest(http.MethodGet, "/debug/pprof/symbol").
		Do().
		Status(http.StatusOK)

	// /debug/vars
	srv.NewRequest(http.MethodGet, "/debug/vars").
		Do().
		Status(http.StatusOK)

	// 命中 h202
	srv.NewRequest(http.MethodGet, "/debug/").
		Do().
		Status(http.StatusAccepted)
}
