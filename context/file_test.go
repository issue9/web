// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

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
	a.NotPanic(func() {
		ctx.ServeFile("./testdata/file", "", map[string]string{"Test": "Test"})
		a.Equal(w.Header().Get("Test"), "Test")
	})

	a.Panic(func() {
		ctx.ServeFile("./testdata/not-exists", "", nil)
	})
}

func TestContext_ServeFileBuffer(t *testing.T) {
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
		ctx.ServeFileBuffer(bytes.NewReader(buf), "name", map[string]string{"Test": "Test"})
		a.Equal(w.Header().Get("Test"), "Test")
	})
}