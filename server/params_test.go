// SPDX-License-Identifier: MIT

package server

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
)

func TestParams_empty(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	server.DefaultRouter().Get("/params/empty", func(ctx *Context) {
		ps := ctx.Params()

		a.Equal(ps.Int64("id1"), 0)
		a.Equal(ps.MustInt64("id2", 2), 2)
		a.True(ps.HasErrors()).
			Equal(1, len(ps.Errors())) // MustInt 在不存在的情况，并不会生成错误信息
	})

	srv := rest.NewServer(t, server.Mux(), nil)
	defer srv.Close()
	srv.Get("/root/params/empty").Do().Status(http.StatusOK)
}

func TestParams_ID_MustID(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	server.DefaultRouter().Get("/params/id/{i1:\\d+}/{i2}/{str}", func(ctx *Context) {
		ps := ctx.Params()

		a.Equal(ps.ID("i1"), 1)

		// 负数
		a.Equal(ps.ID("i2"), -2)
		a.Equal(len(ps.errors), 1)

		// 不存在的参数，添加错误信息
		a.Equal(ps.ID("i3"), 0)
		a.Equal(len(ps.errors), 2)

		// MustID() 不会增加错误信息
		a.Equal(ps.MustID("i3", 3), 3)
		a.Equal(len(ps.errors), 2)

		// MustID() 负数
		a.Equal(ps.MustID("i2", 3), 3)
		a.Equal(len(ps.errors), 2) // 之前已经有一个名为 i2 的错误信息，所以此处为覆盖

		// MustID() 无法转换，会返回默认值，且添加错误信息
		a.Equal(ps.MustID("str", 3), 3)
		a.Equal(len(ps.Errors()), 3)

		// MustID() 能正常转换
		a.Equal(ps.MustID("i1", 3), 1)
		a.Equal(len(ps.Errors()), 3)
	})

	srv := rest.NewServer(t, server.Mux(), nil)
	defer srv.Close()
	srv.Get("/root/params/id/1/-2/str").Do().Status(http.StatusOK)
}

func TestParams_Int_MustInt(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	server.DefaultRouter().Get("/params/int/{i1:\\d+}/{i2:\\d+}/{str}", func(ctx *Context) {
		ps := ctx.Params()

		a.Equal(ps.Int64("i1"), 1)
		a.Equal(ps.Int64("i2"), 2)

		// 不存在的参数，添加错误信息
		a.Equal(ps.Int64("i3"), 0)
		a.Equal(len(ps.errors), 1)

		// MustInt() 不会增加错误信息
		a.Equal(ps.MustInt64("i3", 3), 3)
		a.Equal(len(ps.errors), 1)

		// MustInt() 无法转换，会返回默认值，且添加错误信息
		a.Equal(ps.MustInt64("str", 3), 3)
		a.Equal(len(ps.errors), 2)

		// MustInt() 正常转换
		a.Equal(ps.MustInt64("i1", 3), 1)
		a.Equal(len(ps.errors), 2)
	})

	srv := rest.NewServer(t, server.Mux(), nil)
	defer srv.Close()
	srv.Get("/root/params/int/1/2/str").Do().Status(http.StatusOK)
}

func TestParams_Bool_MustBool(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	server.DefaultRouter().Get("/params/bool/{b1}/{b2}/{str}", func(ctx *Context) {
		ps := ctx.Params()

		a.True(ps.Bool("b1"))
		a.False(ps.Bool("b2"))

		// 不存在的参数，添加错误信息
		a.False(ps.Bool("b3"))
		a.Equal(len(ps.errors), 1)

		// MustBool() 不会增加错误信息
		a.True(ps.MustBool("b3", true))
		a.Equal(len(ps.errors), 1)

		// MustBool() 无法转换，会返回默认值，且添加错误信息
		a.True(ps.MustBool("str", true))
		a.Equal(len(ps.Errors()), 2)

		// MustBool() 能正常转换
		a.True(ps.MustBool("b1", false))
		a.Equal(len(ps.Errors()), 2)
	})

	srv := rest.NewServer(t, server.Mux(), nil)
	defer srv.Close()
	srv.Get("/root/params/bool/true/false/str").Do().Status(http.StatusOK)
}

func TestParams_String_MustString(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	server.DefaultRouter().Get("/params/string/{s1}/{s2}", func(ctx *Context) {
		ps := ctx.Params()

		a.Equal(ps.String("s1"), "str1")
		a.Equal(ps.String("s2"), "str2")
		a.False(ps.HasErrors())

		// 不存在的参数，添加错误信息
		a.Equal(ps.String("s3"), "")
		a.Equal(len(ps.errors), 1)

		// MustString() 不会增加错误信息
		a.Equal(ps.MustString("s3", "str3"), "str3")
		a.Equal(len(ps.Errors()), 1)

		// MustString() 能正常转换
		a.Equal(ps.MustString("s1", "str3"), "str1")
		a.Equal(len(ps.Errors()), 1)
	})

	srv := rest.NewServer(t, server.Mux(), nil)
	defer srv.Close()
	srv.Get("/root/params/string/str1/str2").Do().Status(http.StatusOK)
}

func TestParams_Float_MustFloat(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	server.DefaultRouter().Get("/params/float/{f1}/{f2}/{str}", func(ctx *Context) {
		ps := ctx.Params()

		a.Equal(ps.Float64("f1"), 1.1)
		a.Equal(ps.Float64("f2"), 2.2)
		a.False(ps.HasErrors())

		// 不存在的参数，添加错误信息
		a.Equal(ps.Float64("f3"), 0.0)
		a.Equal(len(ps.errors), 1)

		// MustFloat64() 不会增加错误信息
		a.Equal(ps.MustFloat64("id3", 3.3), 3.3)
		a.Equal(len(ps.errors), 1)

		// MustFloat64() 无法转换，会返回默认值，且添加错误信息
		a.Equal(ps.MustFloat64("str", 3.3), 3.3)
		a.Equal(len(ps.Errors()), 2)

		// MustFloat64() 正常转换
		a.Equal(ps.MustFloat64("f1", 3.3), 1.1)
		a.Equal(len(ps.Errors()), 2)
	})

	srv := rest.NewServer(t, server.Mux(), nil)
	defer srv.Close()
	srv.Get("/root/params/float/1.1/2.2/str").Do().Status(http.StatusOK)
}

func TestContext_ParamID(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	server.DefaultRouter().Get("/params/paramid/{i1}/{i2}/{str}", func(ctx *Context) {
		i1, ok := ctx.ParamID("i1", 41110)
		a.True(ok).Equal(i1, 1)

		i2, ok := ctx.ParamID("i2", 41110)
		a.False(ok).Equal(i2, 0)
	})

	srv := rest.NewServer(t, server.Mux(), nil)
	defer srv.Close()
	srv.Get("/root/params/paramid/1/-2/str").Do().Status(411)
}

func TestContext_ParamInt64(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	server.DefaultRouter().Get("/params/paramint64/{i1}/{i2}/{str}", func(ctx *Context) {
		i1, ok := ctx.ParamInt64("i1", 41110)
		a.True(ok).Equal(i1, 1)

		i2, ok := ctx.ParamInt64("i2", 41110)
		a.True(ok).Equal(i2, -2)

		i3, ok := ctx.ParamInt64("i3", 41110)
		a.False(ok).Equal(i3, 0)
	})

	srv := rest.NewServer(t, server.Mux(), nil)
	defer srv.Close()
	srv.Get("/root/params/paramint64/1/-2/str").Do().Status(411)
}
