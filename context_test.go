// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/web/content"
)

var _ Renderer = &Context{}

var _ Reader = &Context{}

func TestApp_NewContext(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	c, err := content.New(content.DefaultConfig())
	a.NotError(err).NotNil(c)

	ctx := NewContext(w, r, c)
	a.NotNil(ctx)
	a.Equal(ctx.Response(), w).Equal(ctx.Request(), r).Equal(ctx.c, c)
}

func TestContext_ParamID(t *testing.T) {
	a := assert.New(t)

	r, err := http.NewRequest(http.MethodPut, "/test", nil)
	a.NotError(err).NotNil(r)
	ctx := &Context{r: r}
	a.NotNil(ctx)
}

func TestContext_ResultFields(t *testing.T) {
	a := assert.New(t)
	allow := []string{"col1", "col2", "col3"}

	r, err := http.NewRequest(http.MethodPut, "/test", nil)
	a.NotError(err).NotNil(r)
	ctx := &Context{r: r}
	a.NotNil(ctx)

	// 指定的字段都是允许的字段
	ret, ok := ctx.ResultFields(allow)
	a.True(ok).Equal([]string{"col1", "col2", "col3"}, ret)

	// 包含不允许的字段
	r.Header.Set("X-Result-Fields", "col1,col2, col100 ,col101")
	ret, ok = ctx.ResultFields(allow)
	a.False(ok).Equal([]string{"col100", "col101"}, ret)

	// 未指定 X-Result-Fields
	r.Header.Del("X-Result-Fields")
	ret, ok = ctx.ResultFields(allow)
	a.True(ok).Equal(ret, allow)
}
