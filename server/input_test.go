// SPDX-License-Identifier: MIT

package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
	"github.com/issue9/mux/v7"
	"golang.org/x/text/encoding"
	"golang.org/x/text/language"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/serializer/text"
	"github.com/issue9/web/serializer/text/testobject"
)

func newContextWithQuery(a *assert.Assertion, path string) (ctx *Context, w *httptest.ResponseRecorder) {
	r := rest.Post(a, path, []byte("123")).Header("Accept", "*/*").Request()
	w = httptest.NewRecorder()
	ctx = newServer(a, nil).newContext(w, r, nil)

	return ctx, w
}

func TestParams_empty(t *testing.T) {
	a := assert.New(t, false)
	server := newServer(a, nil)
	router := server.Routers().New("default", nil, mux.URLDomain("http://localhost:8081/root"))
	a.NotNil(router)

	router.Get("/params/empty", func(ctx *Context) Responser {
		ps := ctx.Params()

		a.Equal(ps.Int64("id1"), 0)
		a.NotNil(ps.Problem("41110"))
		return nil
	})

	srv := rest.NewServer(a, server.Routers(), nil)
	srv.Get("/params/empty").Do(nil).Status(http.StatusOK)
}

func TestParams_ID(t *testing.T) {
	a := assert.New(t, false)
	server := newServer(a, nil)
	router := server.Routers().New("default", nil, mux.URLDomain("http://localhost:8081/root"))
	a.NotNil(router)

	router.Get("/params/id/{i1:\\d+}/{i2}/{str}", func(ctx *Context) Responser {
		ps := ctx.Params()

		a.Equal(ps.ID("i1"), 1).
			Equal(ps.v.Count(), 0)

		// 负数
		a.Equal(ps.ID("i2"), -2).
			Equal(ps.v.Count(), 1)

		// 不存在的参数，添加错误信息
		a.Equal(ps.ID("i3"), 0)
		a.Equal(ps.v.Count(), 2)

		return nil
	})

	srv := rest.NewServer(a, server.Routers(), nil)
	srv.Get("/params/id/1/-2/str").Do(nil).Status(http.StatusOK)
}

func TestParams_Int(t *testing.T) {
	a := assert.New(t, false)
	server := newServer(a, nil)
	router := server.Routers().New("default", nil, mux.URLDomain("http://localhost:8081/root"))
	a.NotNil(router)

	router.Get("/params/int/{i1:\\d+}/{i2:\\d+}/{str}", func(ctx *Context) Responser {
		ps := ctx.Params()

		a.Equal(ps.Int64("i1"), 1)
		a.Equal(ps.Int64("i2"), 2).Equal(ps.v.Count(), 0)

		// 不存在的参数，添加错误信息
		a.Equal(ps.Int64("i3"), 0)
		a.Equal(ps.v.Count(), 1)

		return nil
	})

	srv := rest.NewServer(a, server.Routers(), nil)
	srv.Get("/params/int/1/2/str").Do(nil).Status(http.StatusOK)
}

func TestParams_Bool(t *testing.T) {
	a := assert.New(t, false)
	server := newServer(a, nil)
	router := server.Routers().New("default", nil, mux.URLDomain("http://localhost:8081"))
	a.NotNil(router)

	router.Get("/params/bool/{b1}/{b2}/{str}", func(ctx *Context) Responser {
		ps := ctx.Params()

		a.True(ps.Bool("b1"))
		a.False(ps.Bool("b2")).Equal(ps.v.Count(), 0)

		// 不存在的参数，添加错误信息
		a.False(ps.Bool("b3"))
		a.Equal(ps.v.Count(), 1)

		return nil
	})

	srv := rest.NewServer(a, server.Routers(), nil)
	srv.Get("/params/bool/true/false/str").Do(nil).Status(http.StatusOK)
}

func TestParams_String(t *testing.T) {
	a := assert.New(t, false)
	server := newServer(a, nil)
	router := server.Routers().New("default", nil, mux.URLDomain("http://localhost:8081/root"))
	a.NotNil(router)

	router.Get("/params/string/{s1}/{s2}", func(ctx *Context) Responser {
		ps := ctx.Params()

		a.Equal(ps.String("s1"), "str1")
		a.Equal(ps.String("s2"), "str2")
		a.Zero(ps.v.Count())

		// 不存在的参数，添加错误信息
		a.Equal(ps.String("s3"), "")
		a.Equal(ps.v.Count(), 1)

		return nil
	})

	srv := rest.NewServer(a, server.Routers(), nil)
	srv.Get("/params/string/str1/str2").Do(nil).Status(http.StatusOK)
}

func TestParams_Float(t *testing.T) {
	a := assert.New(t, false)
	server := newServer(a, nil)
	router := server.Routers().New("default", nil, mux.URLDomain("http://localhost:8081/root"))
	a.NotNil(router)

	router.Get("/params/float/{f1}/{f2}/{str}", func(ctx *Context) Responser {
		ps := ctx.Params()

		a.Equal(ps.Float64("f1"), 1.1)
		a.Equal(ps.Float64("f2"), 2.2)
		a.Zero(ps.v.Count())

		// 不存在的参数，添加错误信息
		a.Equal(ps.Float64("f3"), 0.0)
		a.Equal(ps.v.Count(), 1)

		return nil
	})

	srv := rest.NewServer(a, server.Routers(), nil)
	srv.Get("/params/float/1.1/2.2/str").Do(nil).Status(http.StatusOK)
}

func TestContext_ParamID(t *testing.T) {
	a := assert.New(t, false)
	server := newServer(a, nil)
	router := server.Routers().New("default", nil, mux.URLDomain("http://localhost:8081/root"))
	a.NotNil(router)

	router.Get("/params/paramid/{i1}/{i2}/{str}", func(ctx *Context) Responser {
		i1, resp := ctx.ParamID("i1", "41110")
		a.Nil(resp).Equal(i1, 1)

		i2, resp := ctx.ParamID("i2", "41110")
		a.NotNil(resp).Equal(i2, 0)

		return resp
	})

	srv := rest.NewServer(a, server.Routers(), nil)
	srv.Get("/params/paramid/1/-2/str").Do(nil).Status(411)
}

func TestContext_ParamInt64(t *testing.T) {
	a := assert.New(t, false)
	server := newServer(a, nil)
	router := server.Routers().New("default", nil, mux.URLDomain("http://localhost:8081/root"))
	a.NotNil(router)

	router.Get("/params/paramint64/{i1}/{i2}/{str}", func(ctx *Context) Responser {
		i1, resp := ctx.ParamInt64("i1", "41110")
		a.Nil(resp).Equal(i1, 1)

		i2, resp := ctx.ParamInt64("i2", "41110")
		a.Nil(resp).Equal(i2, -2)

		i3, resp := ctx.ParamInt64("i3", "41110")
		a.NotNil(resp).Equal(i3, 0)

		return resp
	})

	srv := rest.NewServer(a, server.Routers(), nil)
	srv.Get("/params/paramint64/1/-2/str").Do(nil).Status(411)
}

func TestQueries_Int(t *testing.T) {
	a := assert.New(t, false)
	ctx, _ := newContextWithQuery(a, "/queries/int?i1=1&i2=2&str=str")
	q, err := ctx.Queries()
	a.NotError(err).NotNil(q)

	a.Equal(q.Int("i1", 9), 1)
	a.Equal(q.Int("i2", 9), 2)
	a.Equal(q.Int("i3", 9), 9)

	// 无法转换，会返回默认值，且添加错误信息
	a.Equal(q.Int("str", 3), 3)
}

func TestQueries_Int64(t *testing.T) {
	a := assert.New(t, false)
	ctx, _ := newContextWithQuery(a, "/queries/int64?i1=1&i2=2&str=str")
	q, err := ctx.Queries()
	a.NotError(err).NotNil(q)

	a.Equal(q.Int64("i1", 9), 1)
	a.Equal(q.Int64("i2", 9), 2)
	a.Equal(q.Int64("i3", 9), 9)

	// 无法转换，会返回默认值，且添加错误信息
	a.Equal(q.Int64("str", 3), 3)
	a.NotNil(q.Problem("41110"))
}

func TestQueries_String(t *testing.T) {
	a := assert.New(t, false)
	ctx, _ := newContextWithQuery(a, "/queries/string?s1=1&s2=2")
	q, err := ctx.Queries()
	a.NotError(err).NotNil(q)

	a.Equal(q.String("s1", "9"), "1")
	a.Equal(q.String("s2", "9"), "2")
	a.Equal(q.String("s3", "9"), "9")
}

func TestQueries_Bool(t *testing.T) {
	a := assert.New(t, false)
	ctx, _ := newContextWithQuery(a, "/queries/bool?b1=true&b2=true&str=str")
	q, err := ctx.Queries()
	a.NotError(err).NotNil(q)

	a.True(q.Bool("b1", false))
	a.True(q.Bool("b2", false))
	a.False(q.Bool("b3", false))

	// 无法转换，会返回默认值，且添加错误信息
	a.False(q.Bool("str", false))
	a.NotNil(q.Problem("41110"))
}

func TestQueries_Float64(t *testing.T) {
	a := assert.New(t, false)
	ctx, _ := newContextWithQuery(a, "/queries/float64?i1=1.1&i2=2&str=str")
	q, err := ctx.Queries()
	a.NotError(err).NotNil(q)

	a.Equal(q.Float64("i1", 9.9), 1.1)
	a.Equal(q.Float64("i2", 9.9), 2)
	a.Equal(q.Float64("i3", 9.9), 9.9)

	// 无法转换，会返回默认值，且添加错误信息
	a.Equal(q.Float64("str", 3), 3)
	a.NotNil(q.Problem("41110"))
}

func TestContext_Object(t *testing.T) {
	a := assert.New(t, false)
	ctx, w := newContextWithQuery(a, "/queries/float64?i1=1.1&i2=2&str=str")
	q, err := ctx.Queries()
	a.NotError(err).NotNil(q)

	o := struct {
		I1  float32 `query:"i1"`
		I2  int     `query:"i2"`
		Str string  `query:"str"`
	}{}
	resp := ctx.QueryObject(&o, "41110")
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
	resp = ctx.QueryObject(&o2, "41110")
	a.NotNil(resp)
	resp.Apply(ctx)
	a.Equal(w.Code, 411)
}

func TestContext_Body(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, &Options{LanguageTag: language.SimplifiedChinese})

	// 未缓存
	r := rest.Post(a, "/path", []byte("123")).Request()
	w := httptest.NewRecorder()
	ctx := srv.newContext(w, r, nil)
	a.Nil(ctx.body)
	data, err := ctx.Body()
	a.NotError(err).Equal(data, []byte("123"))
	a.Equal(ctx.body, data)

	// 读取缓存内容
	data, err = ctx.Body()
	a.NotError(err).Equal(data, []byte("123"))
	a.Equal(ctx.body, data)

	// 采用 Nop 即 utf-8 编码
	r = rest.Post(a, "/path", []byte("123")).Request()
	ctx = srv.newContext(w, r, nil)
	ctx.inputCharset = encoding.Nop
	data, err = ctx.Body()
	a.NotError(err).Equal(data, []byte("123"))
	a.Equal(ctx.body, data)

	// 采用不同的编码
	r = rest.Post(a, "/path", gbkBytes1).
		Header("Content-type", "text/plain;charset=gb18030").
		Request()
	w = httptest.NewRecorder()
	ctx = srv.newContext(w, r, nil)
	data, err = ctx.Body()
	a.NotError(err).Equal(string(data), gbkString1)
	a.Equal(ctx.body, data)

	// 采用不同的编码
	w = httptest.NewRecorder()
	r = rest.Post(a, "/path", gbkBytes1).
		Header("Accept", "*/*").
		Header("Content-Type", header.BuildContentType(text.Mimetype, " gb18030")).
		Request()
	ctx = srv.newContext(w, r, nil)
	a.NotNil(ctx)
	data, err = ctx.Body()
	a.NotError(err).Equal(string(data), gbkString1)
	a.Equal(ctx.body, data)
}

func TestContext_Unmarshal(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)

	r := rest.Post(a, "/path", []byte("test,123")).
		Header("content-type", text.Mimetype).
		Request()
	w := httptest.NewRecorder()
	ctx := srv.newContext(w, r, nil)

	obj := &testobject.TextObject{}
	a.NotError(ctx.Unmarshal(obj))
	a.Equal(obj.Name, "test").Equal(obj.Age, 123)

	// 无法转换
	o := &struct{}{}
	a.Error(ctx.Unmarshal(o))

	// 空提交
	r = rest.Post(a, "/path", nil).
		Header("content-type", text.Mimetype).
		Request()
	w = httptest.NewRecorder()
	ctx = srv.newContext(w, r, nil)
	obj = &testobject.TextObject{}
	a.NotError(ctx.Unmarshal(obj))
	a.Equal(obj.Name, "").Equal(obj.Age, 0)
}

func TestContext_Read(t *testing.T) {
	a := assert.New(t, false)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", []byte("test,123")).
		Header("Content-Type", header.BuildContentType(text.Mimetype, "utf-8")).
		Request()
	ctx := newServer(a, nil).newContext(w, r, nil)
	obj := &testobject.TextObject{}
	a.Nil(ctx.Read(obj, "41110"))
	a.Equal(obj.Name, "test").Equal(obj.Age, 123)

	w = httptest.NewRecorder()
	r = rest.Post(a, "/path", []byte("test,123")).
		Header("Content-Type", header.BuildContentType(text.Mimetype, "utf-8")).
		Request()
	ctx = newServer(a, nil).newContext(w, r, nil)
	o := &struct{}{}
	resp := ctx.Read(o, "41110")
	a.NotNil(resp)
	resp.Apply(ctx)
	a.Equal(w.Code, http.StatusUnprocessableEntity)
}
