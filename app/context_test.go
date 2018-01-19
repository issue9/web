// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/encoding/simplifiedchinese"

	"github.com/issue9/web/encoding"
)

func TestApp_NewContext(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	w := httptest.NewRecorder()
	app, err := New("./testdata", nil)
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

func TestApp_AddMarshal(t *testing.T) {
	a := assert.New(t)
	f1 := func(v interface{}) ([]byte, error) { return nil, nil }
	app, err := New("./testdata", nil)
	a.NotError(err).NotNil(app)

	a.Equal(1, len(app.marshals))

	a.NotError(app.AddMarshal("n1", f1))
	a.NotError(app.AddMarshal("n2", f1))
	a.Equal(app.AddMarshal("n2", f1), ErrExists)
	a.Equal(3, len(app.marshals))
}

func TestApp_AddUnmarshal(t *testing.T) {
	a := assert.New(t)
	f1 := func(data []byte, v interface{}) error { return nil }
	app, err := New("./testdata", nil)
	a.NotError(err).NotNil(app)

	a.Equal(1, len(app.unmarshals))

	a.NotError(app.AddUnmarshal("n1", f1))
	a.NotError(app.AddUnmarshal("n2", f1))
	a.Equal(app.AddUnmarshal("n2", f1), ErrExists)
	a.Equal(3, len(app.unmarshals))
}

func TestApp_AddCharset(t *testing.T) {
	a := assert.New(t)
	e := simplifiedchinese.GBK
	app, err := New("./testdata", nil)
	a.NotError(err).NotNil(app)

	a.Equal(1, len(app.charset))
	a.Nil(app.charset[encoding.DefaultCharset])

	a.NotError(app.AddCharset("n1", e))
	a.NotError(app.AddCharset("n2", e))
	a.Equal(app.AddCharset("n2", e), ErrExists)
	a.Equal(3, len(app.charset))
}
