// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/issue9/assert"
)

func TestApp_buildHandler(t *testing.T) {
	a := assert.New(t)
	app, err := newApp("./testdata", nil)
	a.NotError(err).NotNil(app)
	m := NewModule("init", "init module")
	m.Get("/ping", h1)
	app.AddModule(m)

	go func() {
		err := app.Run()
		a.ErrorType(err, http.ErrServerClosed, "assert.ErrorType 错误，%v", err.Error())
	}()

	time.Sleep(50 * time.Microsecond)

	resp, err := http.Get("http://localhost:8082/test")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusNotFound)

	// 带正确的域名访问
	resp, err = http.Get("http://127.0.0.1:8082/ping")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)

	// 报头
	resp, err = http.Get("http://localhost:8082/ping")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)
	a.Equal(resp.Header.Get("Access-Control-Allow-Origin"), "*")

	url := app.URL("/debug/pprof/cmdline")
	resp, err = http.Get(url)
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusOK)

	// 命中 h1
	resp, err = http.Get("http://localhost:8082/ping")
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, 1)
}
