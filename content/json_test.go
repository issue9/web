// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package content

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

var _ Content = &json{}

func TestJSON_renderEnvelope(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()
	a.NotNil(w)
	r, err := http.NewRequest("GET", "/index.php?a=b", nil)
	a.NotError(err).NotNil(r)
	j := newJSON(defaultConf)
	a.NotNil(j)

	// 少 Accept
	j.renderEnvelope(w, r, http.StatusOK, nil)
	a.Equal(`{"status":406}`, w.Body.String())
	a.Equal(w.Code, defaultConf.EnvelopeStatus)

	// 错误的
	r.Header.Set("Accept", "test")
	w = httptest.NewRecorder()
	j.renderEnvelope(w, r, http.StatusOK, nil)
	a.Equal(`{"status":406}`, w.Body.String())
	a.Equal(w.Code, defaultConf.EnvelopeStatus)

	// 正常
	r.Header.Set("Accept", jsonContentType)
	w = httptest.NewRecorder()
	w.Header().Set("ContentType", jsonContentType)
	j.renderEnvelope(w, r, http.StatusCreated, nil)
	a.Equal(`{"status":201,"headers":[{"Contenttype":"application/json;charset=utf-8"}]}`, w.Body.String())
	a.Equal(w.Code, defaultConf.EnvelopeStatus)

	// 正常，带 respone
	r.Header.Set("Accept", jsonContentType)
	w = httptest.NewRecorder()
	w.Header().Set("ContentType", jsonContentType)
	j.renderEnvelope(w, r, http.StatusCreated, &struct {
		Field int `json:"field"`
	}{
		Field: 5,
	})
	a.Equal(`{"status":201,"headers":[{"Contenttype":"application/json;charset=utf-8"}],"response":{"field":5}}`, w.Body.String())
	a.Equal(w.Code, defaultConf.EnvelopeStatus)
}

func TestJSON_setHeader(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()
	a.NotNil(w)
	j := newJSON(defaultConf)
	a.NotNil(j)

	j.setHeader(w, nil)
	a.Equal(w.Header().Get("Content-Type"), jsonContentType)

	j.setHeader(w, map[string]string{"Content-Type": "123"})
	a.Equal(w.Header().Get("Content-Type"), "123")
}

func TestJSON_Render(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/index.php?a=b", nil)
	a.NotError(err).NotNil(r)
	j := newJSON(defaultConf)
	a.NotNil(j)

	// 少 accept
	j.Render(w, r, http.StatusCreated, nil, nil)
	a.Equal(w.Code, http.StatusNotAcceptable).Equal(w.Body.String(), "")

	// 错误的 accept
	w = httptest.NewRecorder()
	r.Header.Set("Accept", "test")
	j.Render(w, r, http.StatusCreated, map[string]string{"name": "name"}, map[string]string{"h": "h"})
	a.Equal(w.Code, http.StatusNotAcceptable)
	a.Equal(w.Body.String(), "")

	w = httptest.NewRecorder()
	r.Header.Set("Accept", jsonContentType)
	j.Render(w, r, http.StatusCreated, map[string]string{"name": "name"}, map[string]string{"h": "h"})
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Body.String(), `{"name":"name"}`)
	a.Equal(w.Header().Get("h"), "h")

	// 解析 json 出错，会返回 500 错误
	w = httptest.NewRecorder()
	j.Render(w, r, http.StatusOK, complex(5, 7), nil)
	a.Equal(w.Code, http.StatusInternalServerError)
	a.Equal(w.Body.String(), "")
}

func TestJSON_Read(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()
	a.NotNil(w)
	r, err := http.NewRequest("POST", "/index.php?a=b", bytes.NewBufferString(`{"key":"1"}`))
	a.NotError(err).NotNil(r)
	j := newJSON(defaultConf)
	a.NotNil(j)

	// POST 少 Accept, Content-Type
	val := &struct {
		Key string `json:"key"`
	}{}
	ok := j.Read(w, r, val)
	a.False(ok).
		Equal(w.Code, http.StatusUnsupportedMediaType).
		Equal(val.Key, "")

	// 正常解析
	w = httptest.NewRecorder()
	a.NotNil(w)
	r, err = http.NewRequest("GET", "/index.php?a=b", bytes.NewBufferString(`{"key":"1"}`))
	a.NotError(err).NotNil(r)
	val = &struct {
		Key string `json:"key"`
	}{}
	ok = j.Read(w, r, val)
	a.True(ok).
		Equal(w.Code, http.StatusOK).
		Equal(val.Key, "1")

	// JSON 格式不正确，无法解析
	w = httptest.NewRecorder()
	a.NotNil(w)
	r, err = http.NewRequest("GET", "/index.php?a=b", bytes.NewBufferString(`{"key":1}`))
	a.NotError(err).NotNil(r)
	val = &struct {
		Key string `json:"key"`
	}{}
	ok = j.Read(w, r, val)
	a.False(ok).
		Equal(w.Code, 422).
		Equal(val.Key, "")
}
