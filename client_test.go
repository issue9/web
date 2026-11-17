// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"encoding/json/v2"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/issue9/assert/v6"
	"github.com/issue9/mux/v10/header"

	"github.com/issue9/web/internal/qheader"
	"github.com/issue9/web/problems"
	"github.com/issue9/web/selector"
)

func TestClient_NewRequest(t *testing.T) {
	a := assert.New(t, false)
	nego := newNegotiation(a, nil)

	sel := selector.NewRoundRobin(false, 1)
	sel.Update(selector.NewPeer("https://example.com"))
	c := NewClient(nil, nego, sel, header.JSON, func(v any) ([]byte, error) { return json.Marshal(v) }, header.XRequestID, func() string { return "123" })
	a.NotNil(c).
		NotNil(c.marshal).
		NotNil(c.Client())

	req, err := c.NewRequest(http.MethodPost, "/post", &object{Age: 11})
	a.NotError(err).NotNil(req).
		Equal(req.Header.Get(header.Accept), nego.clientAcceptHeader).
		Equal(req.Header.Get(header.XRequestID), "123").
		Equal(req.Header.Get(header.AcceptEncoding), nego.acceptEncodingHeader).
		Equal(req.Header.Get(header.ContentType), qheader.BuildContentType(header.JSON, header.UTF8))
}

func jsonM(v any) ([]byte, error) { return json.Marshal(v) }

func TestClient_ParseResponse(t *testing.T) {
	a := assert.New(t, false)
	nego := newNegotiation(a, nil)

	sel := selector.NewRoundRobin(false, 1)
	sel.Update(selector.NewPeer("https://example.com"))
	c := NewClient(nil, nego, sel, header.JSON, jsonM, "", nil)
	a.NotNil(c).
		NotNil(c.marshal)

	t.Run("empty", func(*testing.T) {
		resp, problem, err := c.ParseResponse[struct{}, any](&http.Response{})
		a.NotError(err).Nil(problem).Nil(resp)
	})

	t.Run("content-length=0", func(*testing.T) {
		resp := &http.Response{
			Header: http.Header{},
		}
		resp.Header.Set(header.ContentLength, "0")
		res, problem, err := c.ParseResponse[any, any](resp)
		a.NotError(err).Nil(res).Nil(problem)
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

		rsp, problem, err := c.ParseResponse[object, any](resp)
		a.NotError(err).Equal(rsp, obj).Nil(problem)
	})
}

func TestServer_NewClient(t *testing.T) {
	a := assert.New(t, false)

	s := newTestServer(a, false)
	defer s.Run()()
	defer s.Close(500 * time.Millisecond)

	s.Routers().New("default", nil).Get("/get", func(ctx *Context) Responser {
		return OK(&object{Name: "name"})
	}).Post("/post", func(ctx *Context) Responser {
		obj := &object{}
		if resp := ctx.Read(true, obj, problems.BadRequest); resp != nil {
			return resp
		}
		if obj.Name != "name" {
			return ctx.Problem(problems.BadRequest).WithExtensions(&object{Name: "name"})
		}
		return OK(obj)
	})

	sel := selector.NewRoundRobin(false, 1)
	sel.Update(selector.NewPeer(s.st.URL))
	c := s.NewClient(s.st.Client, sel, header.JSON, func(a any) ([]byte, error) { return json.Marshal(a) })
	a.NotNil(c)

	resp, p, err := c.Get[object, any]("/get")
	a.NotError(err).
		Equal(resp, &object{Name: "name"}).
		Nil(p)

	resp, p, err = c.Delete[object, any]("/get")
	a.NotError(err).
		Zero(resp).
		NotNil(p).
		Equal(p.Type, problems.AboutBlank).
		Equal(p.Status, http.StatusMethodNotAllowed)

	resp, p, err = c.Post[object, *object]("/post", nil)
	a.NotError(err).
		Zero(resp).
		NotNil(p).
		Equal(p.Type, problems.AboutBlank).
		Nil(p.Extensions)

	resp, p, err = c.Post[object, *object]("/post", &object{Age: 1, Name: "invalid"})
	a.NotError(err).
		Zero(resp).
		NotNil(p).
		Equal(p.Type, problems.AboutBlank).
		Equal(p.Extensions, &object{Name: "name"})

	resp, p, err = c.Post[object, any]("/post", &object{Age: 1, Name: "name"})
	a.NotError(err).
		Nil(p).
		Equal(resp, &object{Age: 1, Name: "name"})

	resp, p, err = c.Patch[object, any]("/not-exists", nil)
	a.NotError(err).
		Zero(resp).
		NotNil(p).
		Equal(p.Type, problems.AboutBlank)
}
