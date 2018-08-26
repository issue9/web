// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package errors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func clearErrorHandlers() {
	errorHandlers = map[int]func(http.ResponseWriter, int){}
}

func testRender(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	w.Write([]byte("test"))
	w.Header().Set("Content-type", "test")
}

func TestAddErrorHandler(t *testing.T) {
	a := assert.New(t)

	a.NotError(AddErrorHandler(nil, 500, 501))
	a.Error(AddErrorHandler(nil, 500, 502)) // 已经存在

	a.NotError(AddErrorHandler(testRender, 400, 401))
	a.Error(AddErrorHandler(testRender, 401, 402)) // 已经存在

	// 清除内容
	clearErrorHandlers()
}

func TestSetErrorHandler(t *testing.T) {
	a := assert.New(t)

	SetErrorHandler(nil, 500, 501)
	f, found := errorHandlers[500]
	a.True(found).Nil(f)

	SetErrorHandler(testRender, 500, 502)
	a.Equal(errorHandlers[500], testRender)

	// 清除内容
	clearErrorHandlers()
}

func TestRender(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()
	Render(w, http.StatusOK)
	a.Equal(w.Code, http.StatusOK).
		Equal(w.Header().Get("Content-Type"), errorContentType)

	w = httptest.NewRecorder()
	Render(w, http.StatusInternalServerError)
	a.Equal(w.Code, http.StatusInternalServerError).
		Equal(w.Header().Get("Content-Type"), errorContentType)

	// 设置为空，依然采用 defaultRender
	SetErrorHandler(nil, http.StatusInternalServerError)
	w = httptest.NewRecorder()
	Render(w, http.StatusInternalServerError)
	a.Equal(w.Code, http.StatusInternalServerError).
		Equal(w.Header().Get("Content-Type"), errorContentType)

	// 设置为 testRender
	SetErrorHandler(testRender, http.StatusInternalServerError)
	w = httptest.NewRecorder()
	Render(w, http.StatusInternalServerError)
	a.Equal(w.Code, http.StatusInternalServerError).
		Equal(w.Header().Get("Content-Type"), "test").
		Equal(w.Body.String(), "test")

		// 清除内容
	clearErrorHandlers()
}

func TestRender_0(t *testing.T) {
	a := assert.New(t)

	AddErrorHandler(testRender, 401, 402)
	w := httptest.NewRecorder()
	Render(w, 401)
	a.Equal(w.Code, 401).
		Equal(w.Header().Get("Content-Type"), "test").
		Equal(w.Body.String(), "test")
	w = httptest.NewRecorder()
	Render(w, 405) // 不存在
	a.Equal(w.Code, 405).
		Equal(w.Header().Get("Content-Type"), errorContentType)

	// 设置为 testRender
	SetErrorHandler(testRender, 0, 401, 402)
	w = httptest.NewRecorder()
	Render(w, 401)
	a.Equal(w.Code, 401).
		Equal(w.Header().Get("Content-Type"), "test").
		Equal(w.Body.String(), "test")
	w = httptest.NewRecorder()
	Render(w, 405) // 采用 0
	a.Equal(w.Code, 405).
		Equal(w.Header().Get("Content-Type"), "test").
		Equal(w.Body.String(), "test")

		// 清除内容
	clearErrorHandlers()
}
