// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package content

import (
	"bytes"
	stdxml "encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

var _ Content = &xml{}

func TestXML_setHeader(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()
	a.NotNil(w)
	x := newXML(EnvelopeStateDisable, "", http.StatusOK)
	a.NotNil(x)

	x.setHeader(w, nil)
	a.Equal(w.Header().Get("Content-Type"), xmlContentType)

	x.setHeader(w, map[string]string{"Content-Type": "123"})
	a.Equal(w.Header().Get("Content-Type"), "123")
}

func TestXML_Render(t *testing.T) {
	a := assert.New(t)
	buf := new(bytes.Buffer)

	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/index.php?a=b", nil)
	a.NotError(err).NotNil(r)
	x := newXML(EnvelopeStateDisable, "", http.StatusOK)
	a.NotNil(x)

	// 缺少 Accept
	x.Render(w, r, http.StatusCreated, nil, nil)
	a.Equal(w.Code, http.StatusUnsupportedMediaType).Equal(w.Body.String(), "")

	// 错误的 accept
	w = httptest.NewRecorder()
	r.Header.Set("Accept", "test")
	a.NotError(err).NotNil(r)
	x.Render(w, r, http.StatusCreated, nil, nil)
	a.Equal(w.Code, http.StatusUnsupportedMediaType).Equal(w.Body.String(), "")

	val := &struct {
		XMLName stdxml.Name `xml:"root"`
		Name    string      `xml:"name"`
	}{
		Name: "name",
	}
	w = httptest.NewRecorder()
	r.Header.Set("Accept", xmlContentType)
	x.Render(w, r, http.StatusCreated, val, map[string]string{"h": "h"})
	a.Equal(w.Code, http.StatusCreated, buf.String())
	a.Equal(w.Body.String(), `<root><name>name</name></root>`)
	a.Equal(w.Header().Get("h"), "h")

	// 解析xml出错，会返回500错误
	w = httptest.NewRecorder()
	x.Render(w, r, http.StatusOK, complex(5, 7), nil)
	a.Equal(w.Code, http.StatusInternalServerError)
	a.Equal(w.Body.String(), "")
}

func TestXML_Read(t *testing.T) {
	a := assert.New(t)

	// POST 少 Accept, Content-Type
	w := httptest.NewRecorder()
	a.NotNil(w)
	r, err := http.NewRequest("POST", "/index.php?a=b", bytes.NewBufferString(`{"key":"1"}`))
	a.NotError(err).NotNil(r)
	x := newXML(EnvelopeStateDisable, "", http.StatusOK)
	a.NotNil(x)
	val := &struct {
		XMLName stdxml.Name `xml:"root"`
		Key     string      `xml:"key"`
	}{}
	ok := x.Read(w, r, val)
	a.False(ok).
		Equal(w.Code, http.StatusUnsupportedMediaType).
		Equal(val.Key, "")

	// 正常解析
	w = httptest.NewRecorder()
	a.NotNil(w)
	r, err = http.NewRequest("GET", "/index.php?a=b", bytes.NewBufferString(`<root><key>1</key></root>`))
	a.NotError(err).NotNil(r)
	val = &struct {
		XMLName stdxml.Name `xml:"root"`
		Key     string      `xml:"key"`
	}{}
	ok = x.Read(w, r, val)
	a.True(ok).
		Equal(w.Code, http.StatusOK).
		Equal(val.Key, "1")

	// XML 格式不正确，无法解析
	w = httptest.NewRecorder()
	a.NotNil(w)
	r, err = http.NewRequest("GET", "/index.php?a=b", bytes.NewBufferString(`{"key":1}`))
	a.NotError(err).NotNil(r)
	val = &struct {
		XMLName stdxml.Name `xml:"root"`
		Key     string      `xml:"key"`
	}{}
	ok = x.Read(w, r, val)
	a.False(ok).
		Equal(w.Code, http.StatusUnprocessableEntity).
		Equal(val.Key, "")
}
