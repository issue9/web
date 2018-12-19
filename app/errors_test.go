// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/web/internal/exit"
)

func testRenderError(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	w.Write([]byte("test"))
	w.Header().Set("Content-type", "test")
}

func TestApp_AddErrorHandler(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)

	a.NotError(app.AddErrorHandler(nil, 500, 501))
	a.Error(app.AddErrorHandler(nil, 500, 502)) // 已经存在

	a.NotError(app.AddErrorHandler(testRenderError, 400, 401))
	a.Error(app.AddErrorHandler(testRenderError, 401, 402)) // 已经存在
}

func TestApp_SetErrorHandler(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)

	app.SetErrorHandler(nil, 500, 501)
	f, found := app.errorHandlers[500]
	a.True(found).Nil(f)

	app.SetErrorHandler(testRenderError, 500, 502)
	a.Equal(app.errorHandlers[500], ErrorHandler(testRenderError))
}

func TestApp_RenderError(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)

	w := httptest.NewRecorder()
	app.RenderError(w, http.StatusOK)
	a.Equal(w.Code, http.StatusOK)

	w = httptest.NewRecorder()
	app.RenderError(w, http.StatusInternalServerError)
	a.Equal(w.Code, http.StatusInternalServerError)

	// 设置为空，依然采用 defaultRender
	app.SetErrorHandler(nil, http.StatusInternalServerError)
	w = httptest.NewRecorder()
	app.RenderError(w, http.StatusInternalServerError)
	a.Equal(w.Code, http.StatusInternalServerError)

	// 设置为 testRender
	app.SetErrorHandler(testRenderError, http.StatusInternalServerError)
	w = httptest.NewRecorder()
	app.RenderError(w, http.StatusInternalServerError)
	a.Equal(w.Code, http.StatusInternalServerError).
		Equal(w.Header().Get("Content-Type"), "test").
		Equal(w.Body.String(), "test")
}

func TestApp_RenderError_0(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)

	app.AddErrorHandler(testRenderError, 401, 402)
	w := httptest.NewRecorder()
	app.RenderError(w, 401)
	a.Equal(w.Code, 401).
		Equal(w.Header().Get("Content-Type"), "test").
		Equal(w.Body.String(), "test")
	w = httptest.NewRecorder()
	app.RenderError(w, 405) // 不存在
	a.Equal(w.Code, 405)

	// 设置为 testRender
	app.SetErrorHandler(testRenderError, 0, 401, 402)
	w = httptest.NewRecorder()
	app.RenderError(w, 401)
	a.Equal(w.Code, 401).
		Equal(w.Header().Get("Content-Type"), "test").
		Equal(w.Body.String(), "test")
	w = httptest.NewRecorder()
	app.RenderError(w, 405) // 采用 0
	a.Equal(w.Code, 405).
		Equal(w.Header().Get("Content-Type"), "test").
		Equal(w.Body.String(), "test")
}

func TestApp_recovery_debug(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)

	fn := app.recovery(true)

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
	a.NotPanic(func() { fn(w, exit.HTTPStatus(http.StatusBadGateway)) })
	a.Equal(w.Result().StatusCode, http.StatusBadGateway)
	a.True(strings.Contains(w.Body.String(), http.StatusText(http.StatusBadGateway)))
}

func TestApp_recovery(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)
	errLog := new(bytes.Buffer)
	app.ERROR().SetOutput(errLog)

	fn := app.recovery(false)

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
	a.NotPanic(func() { fn(w, exit.HTTPStatus(http.StatusBadGateway)) })
	a.Equal(w.Result().StatusCode, http.StatusBadGateway)
	a.Empty(errLog.String())

	// httpStatus == 0
	w = httptest.NewRecorder()
	errLog.Reset()
	a.NotPanic(func() { fn(w, exit.HTTPStatus(0)) })
	a.Equal(w.Result().StatusCode, http.StatusOK) // 默认输出的状态码
	a.Empty(errLog.String())
}
