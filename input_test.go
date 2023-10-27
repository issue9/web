// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/mux/v7/types"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/internal/testdata"
)

func newContextWithQuery(a *assert.Assertion, path string) (ctx *Context, w *httptest.ResponseRecorder) {
	r := httptest.NewRequest(http.MethodGet, path, bytes.NewBufferString("123"))
	r.Header.Set(header.Accept, "*/*")
	w = httptest.NewRecorder()
	ctx = newTestServer(a).NewContext(w, r)
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

	t.Run("empty", func(t *testing.T) {
		a := assert.New(t, false)
		ctx := NewContext(s, w, r, types.NewContext(), header.RequestIDKey)
		ps := ctx.Paths(false)
		a.Equal(ps.Int64("id1"), 0).
			NotNil(ps.Problem("41110"))
	})

	t.Run("ID", func(*testing.T) {
		ctx := NewContext(s, w, r, newPathContext("i1", "1", "i2", "-2", "i3", "str"), "")
		ps := ctx.Paths(false)

		a.Equal(ps.ID("i1"), 1).
			Equal(ps.filter().len(), 0)

		// 负数
		a.Equal(ps.ID("i2"), 0).
			Equal(ps.filter().len(), 1)

		// 不存在的参数，添加错误信息
		a.Equal(ps.ID("i3"), 0)
		a.Equal(ps.filter().len(), 2)
	})

	t.Run("Int", func(*testing.T) {
		ctx := NewContext(s, w, r, newPathContext("i1", "1", "i2", "-2", "str", "str"), "")
		ps := ctx.Paths(false)

		a.Equal(ps.Int64("i1"), 1)
		a.Equal(ps.Int64("i2"), -2).Equal(ps.filter().len(), 0)

		// 不存在的参数，添加错误信息
		a.Equal(ps.Int64("i3"), 0)
		a.Equal(ps.filter().len(), 1)
	})

	t.Run("Bool", func(*testing.T) {
		ctx := NewContext(s, w, r, newPathContext("b1", "true", "b2", "false", "str", "str"), header.RequestIDKey)
		ps := ctx.Paths(false)

		a.True(ps.Bool("b1"))
		a.False(ps.Bool("b2")).Equal(ps.filter().len(), 0)

		// 不存在的参数，添加错误信息
		a.False(ps.Bool("b3"))
		a.Equal(ps.filter().len(), 1)
	})

	t.Run("String", func(*testing.T) {
		ctx := NewContext(s, w, r, newPathContext("s1", "str1", "s2", "str2"), header.RequestIDKey)
		ps := ctx.Paths(false)

		a.Equal(ps.String("s1"), "str1")
		a.Equal(ps.String("s2"), "str2")
		a.Zero(ps.filter().len())

		// 不存在的参数，添加错误信息
		a.Equal(ps.String("s3"), "")
		a.Equal(ps.filter().len(), 1)
	})

	t.Run("Float", func(*testing.T) {
		ctx := NewContext(s, w, r, newPathContext("f1", "1.1", "f2", "2.2", "str", "str"), header.RequestIDKey)
		ps := ctx.Paths(false)

		a.Equal(ps.Float64("f1"), 1.1)
		a.Equal(ps.Float64("f2"), 2.2)
		a.Zero(ps.filter().len())

		// 不存在的参数，添加错误信息
		a.Equal(ps.Float64("f3"), 0.0)
		a.Equal(ps.filter().len(), 1)
	})
}

func TestContext_PathID(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)

	ctx := NewContext(s, w, r, newPathContext("i1", "1", "i2", "-2", "str", "str"), header.RequestIDKey)

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

	ctx := NewContext(s, w, r, newPathContext("i1", "1", "i2", "-2", "str", "str"), header.RequestIDKey)

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

		a.Equal(q.Int("i1", 9), 1)
		a.Equal(q.Int("i2", 9), 2)
		a.Equal(q.Int("i3", 9), 9)

		// 无法转换，会返回默认值，且添加错误信息
		a.Equal(q.Int("str", 3), 3)
	})

	t.Run("Int64", func(*testing.T) {
		ctx, _ := newContextWithQuery(a, "/queries/int64?i1=1&i2=2&str=str")
		a.NotNil(ctx)
		q, err := ctx.Queries(false)
		a.NotError(err).NotNil(q)

		a.Equal(q.Int64("i1", 9), 1)
		a.Equal(q.Int64("i2", 9), 2)
		a.Equal(q.Int64("i3", 9), 9)

		// 无法转换，会返回默认值，且添加错误信息
		a.Equal(q.Int64("str", 3), 3)
		a.NotNil(q.Problem("41110"))
	})

	t.Run("String", func(*testing.T) {
		ctx, _ := newContextWithQuery(a, "/queries/string?s1=1&s2=2")
		q, err := ctx.Queries(false)
		a.NotError(err).NotNil(q)

		a.Equal(q.String("s1", "9"), "1")
		a.Equal(q.String("s2", "9"), "2")
		a.Equal(q.String("s3", "9"), "9")
	})

	t.Run("Bool", func(*testing.T) {
		ctx, _ := newContextWithQuery(a, "/queries/bool?b1=true&b2=true&str=str")
		q, err := ctx.Queries(false)
		a.NotError(err).NotNil(q)

		a.True(q.Bool("b1", false))
		a.True(q.Bool("b2", false))
		a.False(q.Bool("b3", false))

		// 无法转换，会返回默认值，且添加错误信息
		a.False(q.Bool("str", false))
		a.NotNil(q.Problem("41110"))
	})

	t.Run("Float", func(*testing.T) {
		ctx, _ := newContextWithQuery(a, "/queries/float64?i1=1.1&i2=2&str=str")
		q, err := ctx.Queries(false)
		a.NotError(err).NotNil(q)

		a.Equal(q.Float64("i1", 9.9), 1.1)
		a.Equal(q.Float64("i2", 9.9), 2)
		a.Equal(q.Float64("i3", 9.9), 9.9)

		// 无法转换，会返回默认值，且添加错误信息
		a.Equal(q.Float64("str", 3), 3)
		a.NotNil(q.Problem("41110"))
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

	r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString(testdata.ObjectJSONString))
	r.Header.Set(header.ContentType, "application/json")
	w := httptest.NewRecorder()
	ctx := srv.NewContext(w, r)
	a.NotNil(ctx)

	obj := &testdata.Object{}
	a.NotError(ctx.Unmarshal(obj))
	a.Equal(obj, testdata.ObjectInst)

	// 无法转换
	r = httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString(testdata.ObjectJSONString))
	r.Header.Set(header.ContentType, "application/json")
	w = httptest.NewRecorder()
	ctx = srv.NewContext(w, r)
	a.Error(ctx.Unmarshal(``))

	// 空提交
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set(header.ContentType, "application/json")
	w = httptest.NewRecorder()
	ctx = srv.NewContext(w, r)
	obj = &testdata.Object{}
	a.NotError(ctx.Unmarshal(obj))
	a.Equal(obj.Name, "").Equal(obj.Age, 0)

	// gbk
	r = httptest.NewRequest(http.MethodPost, "/path", bytes.NewBuffer(testdata.ObjectGBKBytes))
	r.Header.Set(header.ContentType, header.BuildContentType("application/json", "gb18030"))
	w = httptest.NewRecorder()
	ctx = srv.NewContext(w, r)
	obj = &testdata.Object{}
	a.NotError(ctx.Unmarshal(obj)).Equal(obj, testdata.ObjectInst)
}

func TestContext_Read(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString(testdata.ObjectJSONString))
	r.Header.Set(header.ContentType, "application/json")
	ctx := s.NewContext(w, r)
	a.NotNil(ctx)
	obj := &testdata.Object{}
	a.Nil(ctx.Read(false, obj, "41110"))
	a.Equal(obj, testdata.ObjectInst)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString(testdata.ObjectJSONString))
	r.Header.Set(header.ContentType, "application/json")
	ctx = s.NewContext(w, r)
	a.NotNil(ctx)
	resp := ctx.Read(false, ``, "41110")
	a.NotNil(resp)
	resp.Apply(ctx)
	a.Equal(w.Code, 422)
}
