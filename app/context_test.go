// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestApp_NewContext(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	w := httptest.NewRecorder()
	conf := defaultConfig()
	app, err := New(conf)
	a.NotError(err).NotNil(app)

	// 缺少 Accept 报头
	app.config.Strict = true
	ctx := app.NewContext(w, r)
	a.Nil(ctx)
	a.Equal(w.Code, http.StatusNotAcceptable)

	// 不检测 Accept 报头
	app.config.Strict = false
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	w = httptest.NewRecorder()
	ctx = app.NewContext(w, r)
	a.NotNil(ctx)
	a.Equal(w.Code, http.StatusOK)
}
