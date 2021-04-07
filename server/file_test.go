// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestContext_ServeFile(t *testing.T) {
	a := assert.New(t)
	b := newServer(a)

	w := httptest.NewRecorder()
	ctx := &Context{
		Response: w,
		Request:  httptest.NewRequest(http.MethodGet, "/file", nil),
		server:   b,
	}

	// 存在的文件
	a.NotPanic(func() {
		ctx.ServeFile("./testdata/file1.txt", map[string]string{"Test": "Test"})
		a.Equal(w.Header().Get("Test"), "Test")
	})

	// 不存在的文件
	w = httptest.NewRecorder()
	ctx.Response = w
	a.NotPanic(func() {
		ctx.ServeFile("./testdata/not-exists", nil)
		a.Equal(w.Code, http.StatusNotFound)
	})
}

func TestContext_ServeContent(t *testing.T) {
	a := assert.New(t)
	b := newServer(a)
	buf, err := ioutil.ReadFile("./testdata/file1.txt")
	a.NotError(err).NotNil(buf)

	w := httptest.NewRecorder()
	ctx := &Context{
		Response: w,
		Request:  httptest.NewRequest(http.MethodGet, "/file1.txt", nil),
		server:   b,
	}

	a.NotPanic(func() {
		ctx.ServeContent(bytes.NewReader(buf), "name", map[string]string{"Test": "Test"})
		a.Equal(w.Header().Get("Test"), "Test")
	})
}
