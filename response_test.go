// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert/v2"
	"github.com/issue9/logs/v4"

	"github.com/issue9/web/serializer/text"
	"github.com/issue9/web/serializer/text/testobject"
	"github.com/issue9/web/server/response"
)

var _ response.Context = &textContext{}

type textContext struct {
	httptest.ResponseRecorder
	l *logs.Logs
}

func newTextContext(a *assert.Assertion) (c *textContext) {
	l := logs.New(logs.NewTextWriter(logs.MicroLayout, &bytes.Buffer{}))
	return &textContext{l: l, ResponseRecorder: *httptest.NewRecorder()}
}

func (c *textContext) Marshal(status int, body any) error {
	data, err := text.Marshal(body)
	if err != nil {
		return err
	}

	c.WriteHeader(status)
	_, err = c.Write(data)
	return err
}

func (c *textContext) Logs() *logs.Logs { return c.l }

func TestCreated(t *testing.T) {
	a := assert.New(t, false)

	ctx := newTextContext(a)
	resp := Created(&testobject.TextObject{Name: "test", Age: 123}, "")
	resp.Apply(ctx)
	a.Equal(ctx.Result().StatusCode, http.StatusCreated).
		Equal(ctx.Body.String(), `test,123`)

	ctx = newTextContext(a)
	resp = Created(&testobject.TextObject{Name: "test", Age: 123}, "/test")
	resp.Apply(ctx)
	a.Equal(ctx.Result().StatusCode, http.StatusCreated).
		Equal(ctx.Body.String(), `test,123`).
		Equal(ctx.Header().Get("Location"), "/test")
}

func TestContext_RetryAfter(t *testing.T) {
	a := assert.New(t, false)

	ctx := newTextContext(a)
	resp := NotImplemented()
	resp.Apply(ctx)
	a.Equal(ctx.Result().StatusCode, http.StatusNotImplemented)

	// Retry-After

	ctx = newTextContext(a)
	resp = RetryAfter(http.StatusServiceUnavailable, 120)
	resp.Apply(ctx)
	a.Equal(ctx.Result().StatusCode, http.StatusServiceUnavailable).
		Empty(ctx.Body.String()).
		Equal(ctx.Header().Get("Retry-After"), "120")

	// Retry-After
	now := time.Now()

	ctx = newTextContext(a)
	resp = RetryAt(http.StatusMovedPermanently, now)
	resp.Apply(ctx)
	a.Equal(ctx.Result().StatusCode, http.StatusMovedPermanently).
		Empty(ctx.Body.String()).
		Contains(ctx.Header().Get("Retry-After"), "GMT")
}

func TestContext_Redirect(t *testing.T) {
	a := assert.New(t, false)

	ctx := newTextContext(a)
	resp := Redirect(301, "https://example.com")
	resp.Apply(ctx)

	a.Equal(ctx.Result().StatusCode, 301).
		Equal(ctx.Header().Get("Location"), "https://example.com")
}
