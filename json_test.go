// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestRenderJSONHeader(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()
	a.NotNil(w)

	renderJSONHeader(w, http.StatusCreated, nil)
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Header().Get("Content-Type"), "application/json;charset=utf-8")

	renderJSONHeader(w, http.StatusCreated, map[string]string{"Content-Type": "123"})
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Header().Get("Content-Type"), "123")
}

func TestRenderJSON(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/index.php?a=b", nil)
	a.NotError(err).NotNil(r)
	RenderJSON(w, r, http.StatusOK, nil, nil)
	a.Equal(w.Code, http.StatusOK).Equal(w.Body.String(), "")

	w = httptest.NewRecorder()
	RenderJSON(w, r, http.StatusInternalServerError, map[string]string{"name": "name"}, map[string]string{"h": "h"})
	a.Equal(w.Code, http.StatusInternalServerError)
	a.Equal(w.Body.String(), `{"name":"name"}`)
	a.Equal(w.Header().Get("h"), "h")

	// 解析json出错，会返回500错误
	w = httptest.NewRecorder()
	RenderJSON(w, r, http.StatusOK, complex(5, 7), nil)
	a.Equal(w.Code, http.StatusInternalServerError)
	a.Equal(w.Body.String(), "")
}
