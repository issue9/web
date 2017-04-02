// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/logs"
)

func newApp(a *assert.Assertion) *App {
	app, err := NewApp("./testdata")
	a.NotError(err).NotNil(app)

	return app
}

func TestApp_File(t *testing.T) {
	a := assert.New(t)

	app := newApp(a)
	a.Equal(app.File("test"), "testdata/test")
	a.Equal(app.File("test/file.jpg"), "testdata/test/file.jpg")
}

func TestNewApp(t *testing.T) {
	a := assert.New(t)

	app := newApp(a)
	a.Equal(app.configDir, "./testdata").
		NotNil(app.config).
		NotNil(app.server).
		NotNil(app.modules).
		Nil(app.handler).
		NotNil(app.content)
}

func TestApp(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)
	logs.SetWriter(logs.LevelError, os.Stderr, "[ERR]", log.LstdFlags)
	logs.SetWriter(logs.LevelInfo, os.Stderr, "[INFO]", log.LstdFlags)

	f1 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(1)
	}
	shutdown := func(w http.ResponseWriter, r *http.Request) {
		app.Shutdown(10 * time.Microsecond)
	}
	restart := func(w http.ResponseWriter, r *http.Request) {
		app.Restart(10 * time.Microsecond)
	}

	app.Mux().GetFunc("/test", f1)
	app.Mux().GetFunc("/shutdown", shutdown)
	app.Mux().GetFunc("/restart", restart)

	go func(app *App) {
		a.NotError(app.Run(nil))
	}(app)

	resp, err := http.Get("http://localhost:8082/test")
	a.NotError(err).NotNil(resp)
}
