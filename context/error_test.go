// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/issue9/assert"
)

var (
	errLog      = new(bytes.Buffer)
	criticalLog = new(bytes.Buffer)
)

func TestContext_Exit(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	app := newApp(a)
	ctx := &Context{
		Response: w,
		App:      app,
	}

	a.Panic(func() {
		ctx.Exit(http.StatusBadRequest)
	})
}

func TestContext_Error(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	app := newApp(a)
	errLog.Reset()
	app.ERROR().SetOutput(errLog)
	ctx := &Context{
		Response: w,
		App:      app,
	}

	ctx.Error(http.StatusInternalServerError, "log1", "log2")
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.HasPrefix(errLog.String(), "log1log2"))
}

func TestContext_Critical(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	criticalLog.Reset()
	app := newApp(a)
	app.CRITICAL().SetOutput(criticalLog)
	ctx := &Context{
		Response: w,
		App:      app,
	}

	ctx.Critical(http.StatusInternalServerError, "log1", "log2")
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.HasPrefix(criticalLog.String(), "log1log2"))
}

func TestContext_Errorf(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	app := newApp(a)
	errLog.Reset()
	app.ERROR().SetOutput(errLog)
	ctx := &Context{
		Response: w,
		App:      app,
	}

	ctx.Errorf(http.StatusInternalServerError, "error @%s:%d", "file.go", 51)
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.HasPrefix(errLog.String(), "error @file.go:51"))
}

func TestContext_Criticalf(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	app := newApp(a)
	criticalLog.Reset()
	app.CRITICAL().SetOutput(criticalLog)
	ctx := &Context{
		Response: w,
		App:      app,
	}

	ctx.Criticalf(http.StatusInternalServerError, "error @%s:%d", "file.go", 51)
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.HasPrefix(criticalLog.String(), "error @file.go:51"))
}
