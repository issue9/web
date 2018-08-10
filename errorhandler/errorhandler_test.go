// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package errorhandler

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

	a.True(AddErrorHandler(500, nil))
	a.False(AddErrorHandler(500, nil)) // 已经存在

	a.True(AddErrorHandler(400, testRender))
	a.False(AddErrorHandler(400, testRender)) // 已经存在

	// 清除内容
	clearErrorHandlers()
}

func TestSetErrorHandler(t *testing.T) {
	a := assert.New(t)

	SetErrorHandler(500, nil)
	f, found := errorHandlers[500]
	a.True(found).Nil(f)

	SetErrorHandler(500, testRender)
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
	SetErrorHandler(http.StatusInternalServerError, nil)
	w = httptest.NewRecorder()
	Render(w, http.StatusInternalServerError)
	a.Equal(w.Code, http.StatusInternalServerError).
		Equal(w.Header().Get("Content-Type"), errorContentType)

	// 设置为 testRender
	SetErrorHandler(http.StatusInternalServerError, testRender)
	w = httptest.NewRecorder()
	Render(w, http.StatusInternalServerError)
	a.Equal(w.Code, http.StatusInternalServerError).
		Equal(w.Header().Get("Content-Type"), "test").
		Equal(w.Body.String(), "test")

		// 清除内容
	clearErrorHandlers()
}
