// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package errorhandler

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

var errLog = new(bytes.Buffer)

func init() {
	logs.SetWriter(logs.LevelError, errLog, "", 0)
}

func TestTraceStack(t *testing.T) {
	a := assert.New(t)

	str := TraceStack(1, "message", 12)
	a.True(strings.HasPrefix(str, "message 12"))
	a.True(strings.Contains(str, "recovery_test.go")) // 肯定包含当前文件名
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
