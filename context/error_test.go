// SPDX-License-Identifier: MIT

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
	b := newServer(a)
	ctx := &Context{
		Response: w,
		server:   b,
	}

	a.Panic(func() {
		ctx.Exit(http.StatusBadRequest)
	})
}

func TestContext_Error(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	b := newServer(a)
	errLog.Reset()
	b.Logs().ERROR().SetOutput(errLog)
	ctx := &Context{
		Response: w,
		server:   b,
	}

	ctx.Error(http.StatusInternalServerError, "log1", "log2")
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.HasPrefix(errLog.String(), "log1log2"))
}

func TestContext_Critical(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	criticalLog.Reset()
	b := newServer(a)
	b.Logs().CRITICAL().SetOutput(criticalLog)
	ctx := &Context{
		Response: w,
		server:   b,
	}

	ctx.Critical(http.StatusInternalServerError, "log1", "log2")
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.HasPrefix(criticalLog.String(), "log1log2"))
}

func TestContext_Errorf(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	b := newServer(a)
	errLog.Reset()
	b.Logs().ERROR().SetOutput(errLog)
	ctx := &Context{
		Response: w,
		server:   b,
	}

	ctx.Errorf(http.StatusInternalServerError, "error @%s:%d", "file.go", 51)
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.HasPrefix(errLog.String(), "error @file.go:51"))
}

func TestContext_Criticalf(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	b := newServer(a)
	criticalLog.Reset()
	b.Logs().CRITICAL().SetOutput(criticalLog)
	ctx := &Context{
		Response: w,
		server:   b,
	}

	ctx.Criticalf(http.StatusInternalServerError, "error @%s:%d", "file.go", 51)
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.HasPrefix(criticalLog.String(), "error @file.go:51"))
}
