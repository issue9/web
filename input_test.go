// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/iotest"

	"github.com/issue9/assert/v3"
	"github.com/issue9/assert/v3/rest"
	"github.com/issue9/mux/v7"
	"golang.org/x/text/language"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/internal/testdata"
	"github.com/issue9/web/servertest"
)

func newContextWithQuery(a *assert.Assertion, path string) (ctx *Context, w *httptest.ResponseRecorder) {
	r := rest.Post(a, path, []byte("123")).Header("Accept", "*/*").Request()
	w = httptest.NewRecorder()
	ctx = newTestServer(a, nil).newContext(w, r, nil)
	return ctx, w
}

func TestPaths_empty(t *testing.T) {
	a := assert.New(t, false)
	server := newTestServer(a, nil)
	router := server.NewRouter("default", nil, mux.URLDomain("http://localhost:8081/root"))
	a.NotNil(router)

	router.Get("/params/empty", func(ctx *Context) Responser {
		ps := ctx.Paths(false)

		a.Equal(ps.Int64("id1"), 0)
		a.NotNil(ps.Problem("41110"))
		return nil
	})

	srv := rest.NewServer(a, server.routers, nil)
	srv.Get("/params/empty").Do(nil).Status(http.StatusOK)
}

func TestPaths_ID(t *testing.T) {
	a := assert.New(t, false)
	server := newTestServer(a, nil)
	router := server.NewRouter("default", nil, mux.URLDomain("http://localhost:8081/root"))
	a.NotNil(router)

	router.Get("/params/id/{i1:\\d+}/{i2}/{str}", func(ctx *Context) Responser {
		ps := ctx.Paths(false)

		a.Equal(ps.ID("i1"), 1).
			Equal(ps.filter.len(), 0)

		// 负数
		a.Equal(ps.ID("i2"), 0).
			Equal(ps.filter.len(), 1)

		// 不存在的参数，添加错误信息
		a.Equal(ps.ID("i3"), 0)
		a.Equal(ps.filter.len(), 2)

		return nil
	})

	srv := rest.NewServer(a, server.routers, nil)
	srv.Get("/params/id/1/-2/str").Do(nil).Status(http.StatusOK)
}

func TestPaths_Int(t *testing.T) {
	a := assert.New(t, false)
	server := newTestServer(a, nil)
	router := server.NewRouter("default", nil, mux.URLDomain("http://localhost:8081/root"))
	a.NotNil(router)

	router.Get("/params/int/{i1:\\d+}/{i2:\\d+}/{str}", func(ctx *Context) Responser {
		ps := ctx.Paths(false)

		a.Equal(ps.Int64("i1"), 1)
		a.Equal(ps.Int64("i2"), 2).Equal(ps.filter.len(), 0)

		// 不存在的参数，添加错误信息
		a.Equal(ps.Int64("i3"), 0)
		a.Equal(ps.filter.len(), 1)

		return nil
	})

	srv := rest.NewServer(a, server.routers, nil)
	srv.Get("/params/int/1/2/str").Do(nil).Status(http.StatusOK)
}

func TestPaths_Bool(t *testing.T) {
	a := assert.New(t, false)
	server := newTestServer(a, nil)
	router := server.NewRouter("default", nil, mux.URLDomain("http://localhost:8081"))
	a.NotNil(router)

	router.Get("/params/bool/{b1}/{b2}/{str}", func(ctx *Context) Responser {
		ps := ctx.Paths(false)

		a.True(ps.Bool("b1"))
		a.False(ps.Bool("b2")).Equal(ps.filter.len(), 0)

		// 不存在的参数，添加错误信息
		a.False(ps.Bool("b3"))
		a.Equal(ps.filter.len(), 1)

		return nil
	})

	srv := rest.NewServer(a, server.routers, nil)
	srv.Get("/params/bool/true/false/str").Do(nil).Status(http.StatusOK)
}

func TestPaths_String(t *testing.T) {
	a := assert.New(t, false)
	server := newTestServer(a, nil)
	router := server.NewRouter("default", nil, mux.URLDomain("http://localhost:8081/root"))
	a.NotNil(router)

	router.Get("/params/string/{s1}/{s2}", func(ctx *Context) Responser {
		ps := ctx.Paths(false)

		a.Equal(ps.String("s1"), "str1")
		a.Equal(ps.String("s2"), "str2")
		a.Zero(ps.filter.len())

		// 不存在的参数，添加错误信息
		a.Equal(ps.String("s3"), "")
		a.Equal(ps.filter.len(), 1)

		return nil
	})

	srv := rest.NewServer(a, server.routers, nil)
	srv.Get("/params/string/str1/str2").Do(nil).Status(http.StatusOK)
}

func TestPaths_Float(t *testing.T) {
	a := assert.New(t, false)
	server := newTestServer(a, nil)
	router := server.NewRouter("default", nil, mux.URLDomain("http://localhost:8081/root"))
	a.NotNil(router)

	router.Get("/params/float/{f1}/{f2}/{str}", func(ctx *Context) Responser {
		ps := ctx.Paths(false)

		a.Equal(ps.Float64("f1"), 1.1)
		a.Equal(ps.Float64("f2"), 2.2)
		a.Zero(ps.filter.len())

		// 不存在的参数，添加错误信息
		a.Equal(ps.Float64("f3"), 0.0)
		a.Equal(ps.filter.len(), 1)

		return nil
	})

	srv := rest.NewServer(a, server.routers, nil)
	srv.Get("/params/float/1.1/2.2/str").Do(nil).Status(http.StatusOK)
}

func TestContext_PathID(t *testing.T) {
	a := assert.New(t, false)
	server := newTestServer(a, nil)
	router := server.NewRouter("default", nil, mux.URLDomain("http://localhost:8081/root"))
	a.NotNil(router)

	router.Get("/params/paramid/{i1}/{i2}/{str}", func(ctx *Context) Responser {
		i1, resp := ctx.PathID("i1", "41110")
		a.Nil(resp).Equal(i1, 1)

		i2, resp := ctx.PathID("i2", "41110")
		a.NotNil(resp).Equal(i2, 0)

		return resp
	})

	srv := rest.NewServer(a, server.routers, nil)
	srv.Get("/params/paramid/1/-2/str").Do(nil).Status(411)
}

func TestContext_PathInt64(t *testing.T) {
	a := assert.New(t, false)
	server := newTestServer(a, nil)
	router := server.NewRouter("default", nil, mux.URLDomain("http://localhost:8081/root"))
	a.NotNil(router)

	router.Get("/params/paramint64/{i1}/{i2}/{str}", func(ctx *Context) Responser {
		i1, resp := ctx.PathInt64("i1", "41110")
		a.Nil(resp).Equal(i1, 1)

		i2, resp := ctx.PathInt64("i2", "41110")
		a.Nil(resp).Equal(i2, -2)

		i3, resp := ctx.PathInt64("i3", "41110")
		a.NotNil(resp).Equal(i3, 0)

		return resp
	})

	srv := rest.NewServer(a, server.routers, nil)
	srv.Get("/params/paramint64/1/-2/str").Do(nil).Status(411)
}

func TestQueries_Int(t *testing.T) {
	a := assert.New(t, false)
	ctx, _ := newContextWithQuery(a, "/queries/int?i1=1&i2=2&str=str")
	q, err := ctx.Queries(false)
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
	q, err := ctx.Queries(false)
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
	q, err := ctx.Queries(false)
	a.NotError(err).NotNil(q)

	a.Equal(q.String("s1", "9"), "1")
	a.Equal(q.String("s2", "9"), "2")
	a.Equal(q.String("s3", "9"), "9")
}

func TestQueries_Bool(t *testing.T) {
	a := assert.New(t, false)
	ctx, _ := newContextWithQuery(a, "/queries/bool?b1=true&b2=true&str=str")
	q, err := ctx.Queries(false)
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
	q, err := ctx.Queries(false)
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

func TestContext_RequestBody(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, &Options{
		Locale:     &Locale{Language: language.SimplifiedChinese},
		HTTPServer: &http.Server{Addr: ":8080"},
	})
	r := srv.NewRouter("def", nil)

	defer servertest.Run(a, srv)()
	defer srv.Close(0)

	// 放第一个，否则 Context.requestBody 一直在复用，无法测试到 content-length == -1 的情况。
	t.Run("content-length=-1", func(t *testing.T) {
		a := assert.New(t, false)

		r.Post("/p4", func(ctx *Context) Responser {
			data, err := ctx.RequestBody()
			a.NotError(err).Equal(data, `"abcdef"`)
			return nil
		})

		reader := iotest.OneByteReader(bytes.NewBufferString(`"abcdef"`))
		resp, err := http.Post("http://localhost:8080/p4", "application/json", reader)
		a.NotError(err).NotNil(resp)
	})

	t.Run("empty", func(t *testing.T) {
		a := assert.New(t, false)
		r.Post("/p1", func(ctx *Context) Responser {
			data, err := ctx.RequestBody()
			a.NotError(err).Empty(data)
			return nil
		})
		servertest.Post(a, "http://localhost:8080/p1", nil).Do(nil).Success()
	})

	t.Run("charset=utf-8", func(t *testing.T) {
		a := assert.New(t, false)
		r.Post("/p2", func(ctx *Context) Responser {
			data, err := ctx.RequestBody()
			a.NotError(err).
				Equal(data, []byte("123")).
				Nil(ctx.inputCharset)

			// 再次读取
			data, err = ctx.RequestBody()
			a.NotError(err).Equal(data, []byte("123"))

			return nil
		})
		servertest.Post(a, "http://localhost:8080/p2", []byte("123")).
			Header("content-type", "application/json").
			Do(nil).Success()
	})

	t.Run("charset=gbk", func(t *testing.T) {
		a := assert.New(t, false)

		r.Post("/p3", func(ctx *Context) Responser {
			data, err := ctx.RequestBody()
			a.NotError(err).Equal(string(data), testdata.ObjectJSONString)
			return nil
		})
		servertest.Post(a, "http://localhost:8080/p3", testdata.ObjectGBKBytes).
			Header("content-type", "application/json;charset=gbk").
			Do(nil).Success()
	})
}

func TestContext_Unmarshal(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, nil)

	r := rest.Post(a, "/path", []byte(testdata.ObjectJSONString)).
		Header("content-type", "application/json").
		Request()
	w := httptest.NewRecorder()
	ctx := srv.newContext(w, r, nil)

	obj := &testdata.Object{}
	a.NotError(ctx.Unmarshal(obj))
	a.Equal(obj, testdata.ObjectInst)

	// 无法转换
	r = rest.Post(a, "/path", []byte(testdata.ObjectJSONString)).
		Header("content-type", "application/json").
		Request()
	w = httptest.NewRecorder()
	ctx = srv.newContext(w, r, nil)
	a.Error(ctx.Unmarshal(``))

	// 空提交
	r = rest.Post(a, "/path", nil).
		Header("content-type", "application/json").
		Request()
	w = httptest.NewRecorder()
	ctx = srv.newContext(w, r, nil)
	obj = &testdata.Object{}
	a.NotError(ctx.Unmarshal(obj))
	a.Equal(obj.Name, "").Equal(obj.Age, 0)
}

func TestContext_Read(t *testing.T) {
	a := assert.New(t, false)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", []byte(testdata.ObjectJSONString)).
		Header("Content-Type", header.BuildContentType("application/json", header.UTF8Name)).
		Request()
	ctx := newTestServer(a, nil).newContext(w, r, nil)
	obj := &testdata.Object{}
	a.Nil(ctx.Read(false, obj, "41110"))
	a.Equal(obj, testdata.ObjectInst)

	w = httptest.NewRecorder()
	r = rest.Post(a, "/path", []byte(testdata.ObjectJSONString)).
		Header("Content-Type", header.BuildContentType("application/json", header.UTF8Name)).
		Request()
	ctx = newTestServer(a, nil).newContext(w, r, nil)
	resp := ctx.Read(false, ``, "41110")
	a.NotNil(resp)
	resp.Apply(ctx)
	a.Equal(w.Code, 422)
}
