// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/internal/header"
)

func TestClient_NewRequest(t *testing.T) {
	a := assert.New(t, false)
	codec := &testCodec{}

	c := NewClient(nil, codec, "application/json", URLSelector("https://example.com"))
	a.NotNil(c).
		NotNil(c.marshal).
		NotNil(c.Client())

	req, err := c.NewRequest(http.MethodPost, "/post", &object{Age: 11})
	a.NotError(err).NotNil(req).
		Equal(req.Header.Get(header.Accept), codec.AcceptHeader()).
		Equal(req.Header.Get(header.AcceptEncoding), codec.AcceptEncodingHeader()).
		Equal(req.Header.Get(header.ContentType), header.BuildContentType("application/json", header.UTF8Name))
}

func TestClient_ParseResponse(t *testing.T) {
	a := assert.New(t, false)
	codec := &testCodec{}

	c := NewClient(nil, codec, "application/json", URLSelector("https://example.com"))
	a.NotNil(c).
		NotNil(c.marshal)

	t.Run("empty", func(*testing.T) {
		resp := &http.Response{}
		p := newRFC7807()
		a.NotError(c.ParseResponse(resp, nil, p))
	})

	t.Run("content-length=0", func(*testing.T) {
		resp := &http.Response{
			Header: http.Header{},
		}
		resp.Header.Set(header.ContentLength, "0")
		p := newRFC7807()
		a.NotError(c.ParseResponse(resp, nil, p))
	})

	t.Run("normal", func(*testing.T) {
		h := http.Header{}
		obj := &object{Age: 11}
		data, err := json.Marshal(obj)
		a.NotError(err).NotNil(data)
		body := bytes.NewBuffer(data)

		h.Set(header.ContentLength, strconv.Itoa(body.Len()))
		h.Set(header.ContentType, header.BuildContentType("application/json", header.UTF8Name))
		h.Set(header.Accept, "application/json")
		h.Set(header.AcceptCharset, header.UTF8Name)
		h.Set(header.AcceptEncoding, "gzip")

		resp := &http.Response{
			Header:     h,
			Body:       io.NopCloser(body),
			StatusCode: http.StatusOK,
		}

		rsp := &object{}
		p := newRFC7807()
		a.NotError(c.ParseResponse(resp, rsp, p))
		a.Equal(rsp, obj).Zero(p)
	})
}
