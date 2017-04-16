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

func TestXML_renderEnvelope(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()
	a.NotNil(w)
	r, err := http.NewRequest("GET", "/index.php?a=b", nil)
	a.NotError(err).NotNil(r)
	j := newXML(defaultConf)
	a.NotNil(j)

	// 少 Accept
	j.renderEnvelope(w, r, http.StatusOK, nil)
	a.Equal(`<xml><status>406</status><headers></headers></xml>`, w.Body.String())
	a.Equal(w.Code, defaultConf.EnvelopeStatus)

	// 错误的
	r.Header.Set("Accept", "test")
	w = httptest.NewRecorder()
	j.renderEnvelope(w, r, http.StatusOK, nil)
	a.Equal(`<xml><status>406</status><headers></headers></xml>`, w.Body.String())
	a.Equal(w.Code, defaultConf.EnvelopeStatus)

	// 正常
	r.Header.Set("Accept", xmlContentType)
	w = httptest.NewRecorder()
	w.Header().Set("Test", "Test")
	j.renderEnvelope(w, r, http.StatusCreated, nil)
	a.Equal(`<xml><status>201</status><headers><header name="Test">Test</header></headers></xml>`, w.Body.String())
	a.Equal(w.Code, defaultConf.EnvelopeStatus)
}

func TestXML_setHeader(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()
	a.NotNil(w)
	x := newXML(defaultConf)
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
	x := newXML(defaultConf)
	a.NotNil(x)

	// 缺少 Accept
	x.Render(w, r, http.StatusCreated, nil, nil)
	a.Equal(w.Code, http.StatusNotAcceptable).
		Equal(w.Body.String(), http.StatusText(http.StatusNotAcceptable)+"\n")

	// 错误的 accept
	w = httptest.NewRecorder()
	r.Header.Set("Accept", "test")
	a.NotError(err).NotNil(r)
	x.Render(w, r, http.StatusCreated, nil, nil)
	a.Equal(w.Code, http.StatusNotAcceptable).
		Equal(w.Body.String(), http.StatusText(http.StatusNotAcceptable)+"\n")

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
	a.Equal(w.Code, http.StatusInternalServerError).
		Equal(w.Body.String(), http.StatusText(http.StatusInternalServerError)+"\n")
}

func TestXML_Read(t *testing.T) {
	a := assert.New(t)

	// POST 少 Accept, Content-Type
	w := httptest.NewRecorder()
	a.NotNil(w)
	r, err := http.NewRequest("POST", "/index.php?a=b", bytes.NewBufferString(`{"key":"1"}`))
	a.NotError(err).NotNil(r)
	x := newXML(defaultConf)
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
