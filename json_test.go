// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"bytes"
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
	a.Equal(w.Header().Get("Content-Type"), "123")
}

func TestRenderJSON(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/index.php?a=b", nil)
	a.NotError(err).NotNil(r)
	RenderJSON(w, r, http.StatusCreated, nil, nil)
	a.Equal(w.Code, http.StatusCreated).Equal(w.Body.String(), "")

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

func TestReadJSON(t *testing.T) {
	a := assert.New(t)

	// POST 少 Accept, Content-Type
	w := httptest.NewRecorder()
	a.NotNil(w)
	r, err := http.NewRequest("POST", "/index.php?a=b", bytes.NewBufferString(`{"key":"1"}`))
	a.NotError(err).NotNil(r)
	val := &struct {
		Key string `json:"key"`
	}{}
	ok := ReadJSON(w, r, val)
	a.False(ok).
		Equal(w.Code, http.StatusUnsupportedMediaType).
		Equal(val.Key, "")

	// 少 Accept
	w = httptest.NewRecorder()
	a.NotNil(w)
	r, err = http.NewRequest("GET", "/index.php?a=b", bytes.NewBufferString(`{"key":"1"}`))
	a.NotError(err).NotNil(r)
	val = &struct {
		Key string `json:"key"`
	}{}
	ok = ReadJSON(w, r, val)
	a.False(ok).
		Equal(w.Code, http.StatusUnsupportedMediaType).
		Equal(val.Key, "")

	// 正常解析
	w = httptest.NewRecorder()
	a.NotNil(w)
	r, err = http.NewRequest("GET", "/index.php?a=b", bytes.NewBufferString(`{"key":"1"}`))
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", "application/json;charset=utf-8")
	val = &struct {
		Key string `json:"key"`
	}{}
	ok = ReadJSON(w, r, val)
	a.True(ok).
		Equal(w.Code, http.StatusOK).
		Equal(val.Key, "1")

	// JSON 格式不正确，无法解析
	w = httptest.NewRecorder()
	a.NotNil(w)
	r, err = http.NewRequest("GET", "/index.php?a=b", bytes.NewBufferString(`{"key":1}`))
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", "application/json;charset=utf-8")
	val = &struct {
		Key string `json:"key"`
	}{}
	ok = ReadJSON(w, r, val)
	a.False(ok).
		Equal(w.Code, 422).
		Equal(val.Key, "")
}
