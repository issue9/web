// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package errors

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/logs"
	"github.com/issue9/web/encoding"
)

var (
	errLog      = new(bytes.Buffer)
	criticalLog = new(bytes.Buffer)
)

func init() {
	logs.SetWriter(logs.LevelError, errLog, "", 0)
	logs.SetWriter(logs.LevelCritical, criticalLog, "", 0)
}

func TestTraceStack(t *testing.T) {
	a := assert.New(t)

	str := traceStack(1, "message", 12)
	a.True(strings.HasPrefix(str, "message12"))
	a.True(strings.Contains(str, "error_test.go")) // 肯定包含当前文件名
}

func TestRenderStatus(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()

	renderStatus(w, http.StatusOK)
	a.Equal(w.Code, http.StatusOK).
		Equal(w.Header().Get("Content-Type"), encoding.BuildContentType(encoding.DefaultMimeType, encoding.DefaultCharset))

	w = httptest.NewRecorder()
	renderStatus(w, http.StatusInternalServerError)
	a.Equal(w.Code, http.StatusInternalServerError).
		Equal(w.Header().Get("Content-Type"), encoding.BuildContentType(encoding.DefaultMimeType, encoding.DefaultCharset))
}

func TestError(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()

	// 没有错误信息
	errLog.Reset()
	Error(2, w, http.StatusInternalServerError)
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(errLog.Len() == 0)

	errLog.Reset()
	Error(2, w, http.StatusInternalServerError, "log1", "log2")
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.HasPrefix(errLog.String(), "log1log2"))
}

func TestCritical(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()

	// 没有错误信息
	criticalLog.Reset()
	Critical(2, w, http.StatusInternalServerError)
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(criticalLog.Len() == 0)

	criticalLog.Reset()
	Critical(2, w, http.StatusInternalServerError, "log1", "log2")
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.HasPrefix(criticalLog.String(), "log1log2"))
}
