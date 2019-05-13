// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
)

var f201 = func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("1234567890"))
}

func TestMiddlewares(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)
	exit := make(chan bool, 1)
	app.errorhandlers.Add(func(w http.ResponseWriter, status int) {
		w.WriteHeader(status)
		w.Write([]byte("error handler test"))
	}, http.StatusNotFound)

	app.Mux().GetFunc("/m1/test", f201)

	go func() {
		err := app.Serve()
		a.ErrorType(err, http.ErrServerClosed, "assert.ErrorType 错误，%v", err.Error())
		exit <- true
	}()
	time.Sleep(500 * time.Microsecond)

	buf := new(bytes.Buffer)
	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8082/m1/test").
		Header("Accept-Encoding", "gzip,deflate;q=0.8").
		Do().
		Status(http.StatusCreated).
		ReadBody(buf).
		Header("Content-Type", "text/html").
		Header("Content-Encoding", "gzip").
		Header("Vary", "Content-Encoding")
	reader, err := gzip.NewReader(buf)
	a.NotError(err).NotNil(reader)
	data, err := ioutil.ReadAll(reader)
	a.NotError(err).NotNil(data)
	a.Equal(string(data), "1234567890")

	// static 中定义的静态文件
	buf.Reset()
	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8082/client/file1.txt").
		Header("Accept-Encoding", "gzip,deflate;q=0.8").
		Do().
		Status(http.StatusOK).
		ReadBody(buf).
		Header("Content-Type", "text/plain; charset=utf-8").
		Header("Content-Encoding", "gzip").
		Header("Vary", "Content-Encoding")
	reader, err = gzip.NewReader(buf)
	a.NotError(err).NotNil(reader)
	data, err = ioutil.ReadAll(reader)
	a.NotError(err).NotNil(data)
	a.Equal(string(data), "file1")

	a.NotError(app.Close())
	<-exit
}

func TestDebug(t *testing.T) {
	srv := rest.NewServer(t, debug(http.HandlerFunc(f202)), nil)
	defer srv.Close()

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
