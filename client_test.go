// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/mux/v9/header"

	"github.com/issue9/web/internal/qheader"
	"github.com/issue9/web/selector"
)

func TestClient_NewRequest(t *testing.T) {
	a := assert.New(t, false)
	codec := newCodec(a)

	sel := selector.NewRoundRobin(false, 1)
	sel.Update(selector.NewPeer("https://example.com"))
	c := NewClient(nil, codec, sel, header.JSON, json.Marshal, header.XRequestID, func() string { return "123" })
	a.NotNil(c).
		NotNil(c.marshal).
		NotNil(c.Client())

	req, err := c.NewRequest(http.MethodPost, "/post", &object{Age: 11})
	a.NotError(err).NotNil(req).
		Equal(req.Header.Get(header.Accept), codec.acceptHeader).
		Equal(req.Header.Get(header.XRequestID), "123").
		Equal(req.Header.Get(header.AcceptEncoding), codec.acceptEncodingHeader).
		Equal(req.Header.Get(header.ContentType), qheader.BuildContentType(header.JSON, header.UTF8))
}

func TestClient_ParseResponse(t *testing.T) {
	a := assert.New(t, false)
	codec := newCodec(a)

	sel := selector.NewRoundRobin(false, 1)
	sel.Update(selector.NewPeer("https://example.com"))
	c := NewClient(nil, codec, sel, header.JSON, json.Marshal, "", nil)
	a.NotNil(c).
		NotNil(c.marshal)

	t.Run("empty", func(*testing.T) {
		resp := &http.Response{}
		p := newProblem
		a.NotError(c.ParseResponse(resp, nil, p))
	})

	t.Run("content-length=0", func(*testing.T) {
		resp := &http.Response{
			Header: http.Header{},
		}
		resp.Header.Set(header.ContentLength, "0")
		a.NotError(c.ParseResponse(resp, nil, nil))
	})

	t.Run("normal", func(*testing.T) {
		h := http.Header{}
		obj := &object{Age: 11}
		data, err := json.Marshal(obj)
		a.NotError(err).NotNil(data)
		body := bytes.NewBuffer(data)

		h.Set(header.ContentLength, strconv.Itoa(body.Len()))
		h.Set(header.ContentType, qheader.BuildContentType(header.JSON, header.UTF8))
		h.Set(header.Accept, header.JSON)
		h.Set(header.AcceptCharset, header.UTF8)
		h.Set(header.AcceptEncoding, "gzip")

		resp := &http.Response{
			Header:        h,
			Body:          io.NopCloser(body),
			StatusCode:    http.StatusOK,
			ContentLength: int64(body.Len()),
		}

		rsp := &object{}
		p := newProblem
		a.NotError(c.ParseResponse(resp, rsp, p)).Equal(rsp, obj)
	})
}
