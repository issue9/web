// SPDX-License-Identifier: MIT

package context

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
	app := newApp(a)

	w := httptest.NewRecorder()
	ctx := &Context{
		Response: w,
		Request:  httptest.NewRequest(http.MethodGet, "/file", nil),
		App:      app,
	}

	// 存在的文件
	a.NotPanic(func() {
		ctx.ServeFile("./testdata/file", map[string]string{"Test": "Test"})
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
	app := newApp(a)
	buf, err := ioutil.ReadFile("./testdata/file")
	a.NotError(err).NotNil(buf)

	w := httptest.NewRecorder()
	ctx := &Context{
		Response: w,
		Request:  httptest.NewRequest(http.MethodGet, "/file", nil),
		App:      app,
	}

	a.NotPanic(func() {
		ctx.ServeContent(bytes.NewReader(buf), "name", map[string]string{"Test": "Test"})
		a.Equal(w.Header().Get("Test"), "Test")
	})
}
