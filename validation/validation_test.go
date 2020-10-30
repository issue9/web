// SPDX-License-Identifier: MIT

package validation

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/cache/memory"
	"github.com/issue9/logs/v2"

	"github.com/issue9/web/context"
	"github.com/issue9/web/context/mimetype/mimetypetest"
)

type (
	root struct {
		O1 *objectWithValidate
		O2 *objectWithoutValidate
	}

	objectWithValidate struct {
		Name string
		Age  int
	}

	objectWithoutValidate struct {
		Name string
		Age  int
	}
)

var (
	_ context.Validator = &root{}
	_ context.Validator = &objectWithValidate{}
)

func (obj *objectWithValidate) Validate(ctx *context.Context) context.ResultFields {
	return New(ctx, ContinueAtError).
		NewField(obj.Age, ".age", Min("不能小于 18", 18)).
		Result()
}

func (root *root) Validate(ctx *context.Context) context.ResultFields {
	return New(ctx, ContinueAtError).
		NewField(root.O1, "o1", If(root.O2 == nil, Required("o1 required", true)).Rules()...).
		NewField(root.O2, "o2", If(root.O1 == nil, Required("o2 required", true)).Rules()...).
		Result()
}

func newContext(a *assert.Assertion) *context.Context {
	u, err := url.Parse("/")
	a.NotError(err).NotNil(u)
	srv := context.NewServer(logs.New(), memory.New(24*time.Hour), false, false, u)
	a.NotNil(srv)
	a.NotError(srv.AddMarshal("text/plain", mimetypetest.TextMarshal))
	a.NotError(srv.AddUnmarshal("text/plain", mimetypetest.TextUnmarshal))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/ctx", nil)
	r.Header.Add("accept", "text/plain")
	ctx := srv.NewContext(w, r)
	a.NotNil(ctx)
	return ctx
}

func TestValidation_ErrorHandling(t *testing.T) {
	a := assert.New(t)
	ctx := newContext(a)

	v := New(ctx, ContinueAtError).
		NewField(-100, "f1", Min("-2", -2), Min("-3", -3)).
		NewField(100, "f2", Max("50", 50), Max("-4", -4))
	a.Equal(v.Result(), map[string][]string{
		"f1": {"-2", "-3"},
		"f2": {"50", "-4"},
	})

	v = New(ctx, ExitFieldAtError).
		NewField(-100, "f1", Min("-2", -2), Min("-3", -3)).
		NewField(100, "f2", Max("50", 50), Max("-4", -4))
	a.Equal(v.Result(), map[string][]string{
		"f1": {"-2"},
		"f2": {"50"},
	})

	v = New(ctx, ExitAtError).
		NewField(-100, "f1", Min("-2", -2), Min("-3", -3)).
		NewField(100, "f2", Max("50", 50), Max("-4", -4))
	a.Equal(v.Result(), map[string][]string{
		"f1": {"-2"},
	})
}

func TestValidation_NewObject(t *testing.T) {
	a := assert.New(t)

	ctx := newContext(a)
	obj := &objectWithValidate{}
	v := New(ctx, ContinueAtError).
		NewField(obj, "obj")
	a.Equal(v.Result(), map[string][]string{
		"obj.age": {"不能小于 18"},
	})

	//
	ctx = newContext(a)
	r := root{}
	errs := r.Validate(ctx)
	a.Equal(errs, map[string][]string{
		"o1": {"o1 required"},
		"o2": {"o2 required"},
	})

	ctx = newContext(a)
	r = root{O1: &objectWithValidate{}}
	errs = r.Validate(ctx)
	a.Equal(errs, map[string][]string{
		"o1.age": {"不能小于 18"},
	})
}
