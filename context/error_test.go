// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/logs"
)

var (
	errLog      = new(bytes.Buffer)
	criticalLog = new(bytes.Buffer)
)

func init() {
	logs.SetWriter(logs.LevelError, errLog, "", 0)
	logs.SetWriter(logs.LevelCritical, criticalLog, "", 0)
}

func TestContext_Exit(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	ctx := &Context{Response: w}

	a.Panic(func() {
		ctx.Exit(http.StatusBadRequest)
	})
}

func TestContext_Error(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	ctx := &Context{Response: w}

	errLog.Reset()
	ctx.Error(http.StatusInternalServerError, "log1", "log2")
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.HasPrefix(errLog.String(), "log1 log2"))
}

func TestContext_Critical(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	ctx := &Context{Response: w}

	criticalLog.Reset()
	ctx.Critical(http.StatusInternalServerError, "log1", "log2")
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.HasPrefix(criticalLog.String(), "log1 log2"))
}

func TestContext_Errorf(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	ctx := &Context{Response: w}

	errLog.Reset()
	ctx.Errorf(http.StatusInternalServerError, "error @%s:%d", "file.go", 51)
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.HasPrefix(errLog.String(), "error @file.go:51"))
}

func TestContext_Criticalf(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	ctx := &Context{Response: w}

	criticalLog.Reset()
	ctx.Criticalf(http.StatusInternalServerError, "error @%s:%d", "file.go", 51)
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.HasPrefix(criticalLog.String(), "error @file.go:51"))
}

func TestError(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()

	errLog.Reset()
	Error(w, http.StatusInternalServerError, "log1", "log2")
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.HasPrefix(errLog.String(), "log1 log2"))
}

func TestCritical(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()

	criticalLog.Reset()
	Critical(w, http.StatusInternalServerError, "log1", "log2")
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.HasPrefix(criticalLog.String(), "log1 log2"))
}

func TestErrorf(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()

	errLog.Reset()
	Errorf(w, http.StatusInternalServerError, "error @%s:%d", "file.go", 51)
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.HasPrefix(errLog.String(), "error @file.go:51"))
}

func TestCriticalf(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()

	criticalLog.Reset()
	Criticalf(w, http.StatusInternalServerError, "error @%s:%d", "file.go", 51)
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.HasPrefix(criticalLog.String(), "error @file.go:51"))
}

func TestTraceStack(t *testing.T) {
	a := assert.New(t)

	str := traceStack(1, "message", 12)
	a.True(strings.HasPrefix(str, "message 12"))
	a.True(strings.Contains(str, "error_test.go")) // 肯定包含当前文件名
}

func TestRecovery_debug(t *testing.T) {
	a := assert.New(t)
	fn := Recovery(true)

	w := httptest.NewRecorder()
	a.NotPanic(func() { fn(w, "msg") })

	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.Contains(w.Body.String(), "msg"))

	// 普通数值
	w = httptest.NewRecorder()
	a.NotPanic(func() { fn(w, http.StatusBadGateway) })

	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.Contains(w.Body.String(), strconv.FormatInt(http.StatusBadGateway, 10)))

	// httpStatus
	w = httptest.NewRecorder()
	a.NotPanic(func() { fn(w, httpStatus(http.StatusBadGateway)) })
	a.Equal(w.Result().StatusCode, http.StatusBadGateway)
	a.True(strings.Contains(w.Body.String(), http.StatusText(http.StatusBadGateway)))
}

func TestRecovery(t *testing.T) {
	a := assert.New(t)
	fn := Recovery(false)

	w := httptest.NewRecorder()
	a.NotPanic(func() { fn(w, "msg") })
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.Contains(errLog.String(), "msg"))

	// 普通数值
	w = httptest.NewRecorder()
	errLog.Reset()
	a.NotPanic(func() { fn(w, http.StatusBadGateway) })
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.Contains(errLog.String(), strconv.FormatInt(http.StatusBadGateway, 10)))

	// httpStatus
	w = httptest.NewRecorder()
	errLog.Reset()
	a.NotPanic(func() { fn(w, httpStatus(http.StatusBadGateway)) })
	a.Equal(w.Result().StatusCode, http.StatusBadGateway)
	a.Empty(errLog.String())

	// httpStatus == 0
	w = httptest.NewRecorder()
	errLog.Reset()
	a.NotPanic(func() { fn(w, httpStatus(0)) })
	a.Equal(w.Result().StatusCode, http.StatusOK) // 默认输出的状态码
	a.Empty(errLog.String())
}
