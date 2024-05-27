// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/mux/v9/header"
	"github.com/issue9/mux/v9/types"

	"github.com/issue9/web/internal/qheader"
)

var (
	objectInst = &object{
		Name: "中文",
		Age:  456,
	}

	// {"name":"中文","Age":456}
	objectGBKBytes = []byte{'{', '"', 'n', 'a', 'm', 'e', '"', ':', '"', 214, 208, 206, 196, '"', ',', '"', 'A', 'g', 'e', '"', ':', '4', '5', '6', '}'}
)

const objectJSONString = `{"name":"中文","Age":456}`

type object struct {
	Name string `json:"name"`
	Age  int
}

func newContextWithQuery(a *assert.Assertion, path string) (ctx *Context, w *httptest.ResponseRecorder) {
	r := httptest.NewRequest(http.MethodGet, path, bytes.NewBufferString("123"))
	r.Header.Set(header.Accept, "*/*")
	w = httptest.NewRecorder()
	ctx = newTestServer(a).NewContext(w, r, types.NewContext())
	return ctx, w
}

func newPathContext(kv ...string) *types.Context {
	c := types.NewContext()
	for i := 0; i < len(kv); i += 2 {
		c.Set(kv[i], kv[i+1])
	}
	return c
}

func TestPaths(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	b := s.InternalServer

	t.Run("empty", func(t *testing.T) {
		a := assert.New(t, false)
		ctx := b.NewContext(w, r, types.NewContext())
		ps := ctx.Paths(false)
		a.Equal(ps.Int64("id1"), 0).
			NotNil(ps.Problem("41110"))
	})

	t.Run("ID", func(*testing.T) {
		ctx := b.NewContext(w, r, newPathContext("i1", "1", "i2", "-2", "i3", "str"))
		ps := ctx.Paths(false)

		a.Equal(ps.ID("i1"), 1).
			Equal(ps.filter().len(), 0)

		// 负数
		a.Equal(ps.ID("i2"), 0).
			Equal(ps.filter().len(), 1)

		// 不存在的参数，添加错误信息
		a.Equal(ps.ID("i3"), 0).
			Equal(ps.filter().len(), 2)
	})

	t.Run("Int", func(*testing.T) {
		ctx := b.NewContext(w, r, newPathContext("i1", "1", "i2", "-2", "str", "str"))
		ps := ctx.Paths(false)

		a.Equal(ps.Int64("i1"), 1).
			Equal(ps.Int64("i2"), -2).Equal(ps.filter().len(), 0)

		// 不存在的参数，添加错误信息
		a.Equal(ps.Int64("i3"), 0).
			Equal(ps.filter().len(), 1)
	})

	t.Run("Bool", func(*testing.T) {
		ctx := b.NewContext(w, r, newPathContext("b1", "true", "b2", "false", "str", "str"))
		ps := ctx.Paths(false)

		a.True(ps.Bool("b1")).False(ps.Bool("b2")).Equal(ps.filter().len(), 0)

		// 不存在的参数，添加错误信息
		a.False(ps.Bool("b3")).
			Equal(ps.filter().len(), 1)
	})

	t.Run("String", func(*testing.T) {
		ctx := b.NewContext(w, r, newPathContext("s1", "str1", "s2", "str2"))
		ps := ctx.Paths(false)

		a.Equal(ps.String("s1"), "str1").
			Equal(ps.String("s2"), "str2").
			Zero(ps.filter().len())

		// 不存在的参数，添加错误信息
		a.Equal(ps.String("s3"), "").
			Equal(ps.filter().len(), 1)
	})

	t.Run("Float", func(*testing.T) {
		ctx := b.NewContext(w, r, newPathContext("f1", "1.1", "f2", "2.2", "str", "str"))
		ps := ctx.Paths(false)

		a.Equal(ps.Float64("f1"), 1.1).
			Equal(ps.Float64("f2"), 2.2).
			Zero(ps.filter().len())

		// 不存在的参数，添加错误信息
		a.Equal(ps.Float64("f3"), 0.0).
			Equal(ps.filter().len(), 1)
	})
}

func TestContext_PathID(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	b := s.InternalServer

	ctx := b.NewContext(w, r, newPathContext("i1", "1", "i2", "-2", "str", "str"))

	i1, resp := ctx.PathID("i1", "41110")
	a.Nil(resp).Equal(i1, 1)

	i2, resp := ctx.PathID("i2", "41110")
	a.NotNil(resp).Equal(i2, 0)
}

func TestContext_PathInt64(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	b := s.InternalServer

	ctx := b.NewContext(w, r, newPathContext("i1", "1", "i2", "-2", "str", "str"))

	i1, resp := ctx.PathInt64("i1", "41110")
	a.Nil(resp).Equal(i1, 1)

	i2, resp := ctx.PathInt64("i2", "41110")
	a.Nil(resp).Equal(i2, -2)

	i3, resp := ctx.PathInt64("i3", "41110")
	a.NotNil(resp).Equal(i3, 0)
}

func TestQueries(t *testing.T) {
	a := assert.New(t, false)

	t.Run("Int", func(*testing.T) {
		ctx, _ := newContextWithQuery(a, "/queries/int?i1=1&i2=2&str=str")
		a.NotNil(ctx)
		q, err := ctx.Queries(false)
		a.NotError(err).NotNil(q)

		a.Equal(q.Int("i1", 9), 1).
			Equal(q.Int("i2", 9), 2).
			Equal(q.Int("i3", 9), 9)

		// 无法转换，会返回默认值，且添加错误信息
		a.Equal(q.Int("str", 3), 3)
	})

	t.Run("Int64", func(*testing.T) {
		ctx, _ := newContextWithQuery(a, "/queries/int64?i1=1&i2=2&str=str")
		a.NotNil(ctx)
		q, err := ctx.Queries(false)
		a.NotError(err).NotNil(q)

		a.Equal(q.Int64("i1", 9), 1).
			Equal(q.Int64("i2", 9), 2).
			Equal(q.Int64("i3", 9), 9)

		// 无法转换，会返回默认值，且添加错误信息
		a.Equal(q.Int64("str", 3), 3).
			NotNil(q.Problem("41110"))
	})

	t.Run("String", func(*testing.T) {
		ctx, _ := newContextWithQuery(a, "/queries/string?s1=1&s2=2")
		q, err := ctx.Queries(false)
		a.NotError(err).NotNil(q)

		a.Equal(q.String("s1", "9"), "1").
			Equal(q.String("s2", "9"), "2").
			Equal(q.String("s3", "9"), "9")
	})

	t.Run("Bool", func(*testing.T) {
		ctx, _ := newContextWithQuery(a, "/queries/bool?b1=true&b2=true&str=str")
		q, err := ctx.Queries(false)
		a.NotError(err).NotNil(q)

		a.True(q.Bool("b1", false)).
			True(q.Bool("b2", false)).
			False(q.Bool("b3", false))

		// 无法转换，会返回默认值，且添加错误信息
		a.False(q.Bool("str", false)).NotNil(q.Problem("41110"))
	})

	t.Run("Float", func(*testing.T) {
		ctx, _ := newContextWithQuery(a, "/queries/float64?i1=1.1&i2=2&str=str")
		q, err := ctx.Queries(false)
		a.NotError(err).NotNil(q)

		a.Equal(q.Float64("i1", 9.9), 1.1).
			Equal(q.Float64("i2", 9.9), 2).
			Equal(q.Float64("i3", 9.9), 9.9)

		// 无法转换，会返回默认值，且添加错误信息
		a.Equal(q.Float64("str", 3), 3).NotNil(q.Problem("41110"))
	})
}

func TestContext_Object(t *testing.T) {
	a := assert.New(t, false)
	ctx, w := newContextWithQuery(a, "/queries/float64?i1=1.1&i2=2&str=str")
	q, err := ctx.Queries(false)
	a.NotError(err).NotNil(q)

	o := struct {
		I1  float32 `query:"i1"`
		I2  int     `query:"i2"`
		Str string  `query:"str"`
	}{}
	resp := ctx.QueryObject(false, &o, "41110")
	a.Nil(resp).
		Equal(w.Code, http.StatusOK)
	a.Equal(o.I1, float32(1.1)).
		Equal(o.I2, 2).
		Equal(o.Str, "str")

	o2 := struct {
		I1  float32 `query:"i1"`
		I2  int     `query:"i2"`
		Str int     `query:"str"`
	}{}
	resp = ctx.QueryObject(false, &o2, "41110")
	a.NotNil(resp)
	resp.Apply(ctx)
	a.Equal(w.Code, 411)
}

func TestContext_Unmarshal(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a)

	r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString(objectJSONString))
	r.Header.Set(header.ContentType, header.JSON)
	w := httptest.NewRecorder()
	ctx := srv.NewContext(w, r, types.NewContext())
	a.NotNil(ctx)

	obj := &object{}
	a.NotError(ctx.Unmarshal(obj)).Equal(obj, objectInst)

	// 无法转换
	r = httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString(objectJSONString))
	r.Header.Set(header.ContentType, header.JSON)
	w = httptest.NewRecorder()
	ctx = srv.NewContext(w, r, types.NewContext())
	a.Error(ctx.Unmarshal(``))

	// 空提交
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set(header.ContentType, header.JSON)
	w = httptest.NewRecorder()
	ctx = srv.NewContext(w, r, types.NewContext())
	obj = &object{}
	a.NotError(ctx.Unmarshal(obj))
	a.Equal(obj.Name, "").Equal(obj.Age, 0)

	// gbk
	r = httptest.NewRequest(http.MethodPost, "/path", bytes.NewBuffer(objectGBKBytes))
	r.Header.Set(header.ContentType, qheader.BuildContentType(header.JSON, "gb18030"))
	w = httptest.NewRecorder()
	ctx = srv.NewContext(w, r, types.NewContext())
	obj = &object{}
	a.NotError(ctx.Unmarshal(obj)).Equal(obj, objectInst)
}

func TestContext_Read(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString(objectJSONString))
	r.Header.Set(header.ContentType, header.JSON)
	ctx := s.NewContext(w, r, types.NewContext())
	a.NotNil(ctx)
	obj := &object{}
	a.Nil(ctx.Read(false, obj, "41110"))
	a.Equal(obj, objectInst)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString(objectJSONString))
	r.Header.Set(header.ContentType, header.JSON)
	ctx = s.NewContext(w, r, types.NewContext())
	a.NotNil(ctx)
	resp := ctx.Read(false, ``, "41110")
	a.NotNil(resp)
	resp.Apply(ctx)
	a.Equal(w.Code, 422)
}
