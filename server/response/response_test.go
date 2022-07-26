// SPDX-License-Identifier: MIT

package response

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/logs/v4"
)

var _ Context = &jsonContext{}

type jsonContext struct {
	httptest.ResponseRecorder
	l *logs.Logs
}

func newJSONContext(a *assert.Assertion) (c *jsonContext, errlog *bytes.Buffer) {
	errlog = &bytes.Buffer{}
	l := logs.New(logs.NewTextWriter(logs.MicroLayout, errlog))
	return &jsonContext{l: l, ResponseRecorder: *httptest.NewRecorder()}, errlog
}

func (c *jsonContext) Marshal(status int, body any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}

	c.WriteHeader(status)
	_, err = c.Write(data)
	return err
}

func (c *jsonContext) Logs() *logs.Logs { return c.l }

func TestStatus(t *testing.T) {
	a := assert.New(t, false)

	a.PanicString(func() {
		Status(http.StatusOK, "k", "v", "K2")
	}, "kv 必须偶数位")

	resp := Status(http.StatusAccepted)
	a.NotNil(resp)
	ctx, errlog := newJSONContext(a)
	resp.Apply(ctx)
	a.Equal(ctx.Result().StatusCode, http.StatusAccepted).Empty(errlog.String())

	resp = Status(http.StatusAccepted, "k1", "v1", "k2", "v2")
	a.NotNil(resp)
	ctx, errlog = newJSONContext(a)
	resp.Apply(ctx)
	a.Equal(ctx.Result().StatusCode, http.StatusAccepted).
		Empty(errlog.String()).
		Equal(ctx.Header().Get("k1"), "v1").
		Equal(ctx.Header().Get("k2"), "v2")
}

func TestObject(t *testing.T) {
	a := assert.New(t, false)

	a.PanicString(func() {
		Object(http.StatusOK, nil, "k", "v", "K2")
	}, "kv 必须偶数位")

	resp := Object(http.StatusAccepted, "obj")
	a.NotNil(resp)
	ctx, errlog := newJSONContext(a)
	resp.Apply(ctx)
	a.Equal(ctx.Result().StatusCode, http.StatusAccepted).
		Empty(errlog.String()).
		Equal(ctx.Body.String(), `"obj"`)

	resp = Object(http.StatusAccepted, struct{ K1 string }{K1: "V1"}, "k1", "v1", "k2", "v2")
	a.NotNil(resp)
	ctx, errlog = newJSONContext(a)
	resp.Apply(ctx)
	a.Equal(ctx.Result().StatusCode, http.StatusAccepted).
		Empty(errlog.String()).
		Equal(ctx.Header().Get("k1"), "v1").
		Equal(ctx.Header().Get("k2"), "v2").
		Equal(ctx.Body.String(), `{"K1":"V1"}`)

	resp = Object(http.StatusAccepted, make(chan int))
	a.NotNil(resp)
	ctx, errlog = newJSONContext(a)
	resp.Apply(ctx)
	a.NotEmpty(errlog.String()).
		Empty(ctx.Body.String())
}
