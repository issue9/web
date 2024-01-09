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
	"github.com/issue9/web/selector"
)

func TestClient_NewRequest(t *testing.T) {
	a := assert.New(t, false)
	codec := newCodec(a)

	sel := selector.NewRoundRobin(false, 1)
	a.NotError(sel.Add(selector.NewPeer("https://example.com")))
	c := NewClient(nil, codec, sel, "application/json", json.Marshal, header.RequestIDKey, func() string { return "123" })
	a.NotNil(c).
		NotNil(c.marshal).
		NotNil(c.Client())

	req, err := c.NewRequest(http.MethodPost, "/post", &object{Age: 11})
	a.NotError(err).NotNil(req).
		Equal(req.Header.Get(header.Accept), codec.acceptHeader).
		Equal(req.Header.Get(header.RequestIDKey), "123").
		Equal(req.Header.Get(header.AcceptEncoding), codec.acceptEncodingHeader).
		Equal(req.Header.Get(header.ContentType), header.BuildContentType("application/json", header.UTF8Name))
}

func TestClient_ParseResponse(t *testing.T) {
	a := assert.New(t, false)
	codec := newCodec(a)

	sel := selector.NewRoundRobin(false, 1)
	a.NotError(sel.Add(selector.NewPeer("https://example.com")))
	c := NewClient(nil, codec, sel, "application/json", json.Marshal, "", nil)
	a.NotNil(c).
		NotNil(c.marshal)

	t.Run("empty", func(*testing.T) {
		resp := &http.Response{}
		p := newProblem()
		a.NotError(c.ParseResponse(resp, nil, p))
	})

	t.Run("content-length=0", func(*testing.T) {
		resp := &http.Response{
			Header: http.Header{},
		}
		resp.Header.Set(header.ContentLength, "0")
		p := newProblem()
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
			Header:        h,
			Body:          io.NopCloser(body),
			StatusCode:    http.StatusOK,
			ContentLength: int64(body.Len()),
		}

		rsp := &object{}
		p := newProblem()
		a.NotError(c.ParseResponse(resp, rsp, p))
		a.Equal(rsp, obj).Zero(p)
	})
}
