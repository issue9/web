// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/mux"

	"github.com/issue9/web/encoding"
)

func newContextWithParam(w http.ResponseWriter, r *http.Request, a *assert.Assertion) *Context {
	r.Header.Set("Accept", "*/*")
	ctx := newContext(w, r, encoding.TextMarshal, nil, encoding.TextUnmarshal, nil)

	return ctx
}

func TestParams_empty(t *testing.T) {
	a := assert.New(t)
	mux := mux.New(true, true, nil, nil)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.GetFunc("/params/empty", func(w http.ResponseWriter, r *http.Request) {
		ctx := newContextWithParam(w, r, a)
		ps := ctx.Params()

		a.Equal(ps.Int("id1"), 0)
		a.Equal(ps.MustInt("id2", 2), 2)
		a.True(ps.HasErrors()).
			Equal(1, len(ps.Errors())) // MustInt 在不存在的情况，并不会生成错误信息
	})

	resp, err := http.Get(srv.URL + "/params/empty")
	a.NotError(err).NotNil(resp)
}

func TestParams_ID_MustID(t *testing.T) {
	a := assert.New(t)
	mux := mux.New(true, true, nil, nil)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.GetFunc("/params/id/{i1:\\d+}/{i2}/{str}", func(w http.ResponseWriter, r *http.Request) {
		ctx := newContextWithParam(w, r, a)
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

	resp, err := http.Get(srv.URL + "/params/id/1/-2/str")
	a.NotError(err).NotNil(resp)
}

func TestParams_Int_MustInt(t *testing.T) {
	a := assert.New(t)
	mux := mux.New(true, true, nil, nil)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.GetFunc("/params/int/{i1:\\d+}/{i2:\\d+}/{str}", func(w http.ResponseWriter, r *http.Request) {
		ctx := newContextWithParam(w, r, a)
		ps := ctx.Params()

		a.Equal(ps.Int("i1"), 1)
		a.Equal(ps.Int("i2"), 2)

		// 不存在的参数，添加错误信息
		a.Equal(ps.Int("i3"), 0)
		a.Equal(len(ps.errors), 1)

		// MustInt() 不会增加错误信息
		a.Equal(ps.MustInt("i3", 3), 3)
		a.Equal(len(ps.errors), 1)

		// MustInt() 无法转换，会返回默认值，且添加错误信息
		a.Equal(ps.MustInt("str", 3), 3)
		a.Equal(len(ps.errors), 2)

		// MustInt() 正常转换
		a.Equal(ps.MustInt("i1", 3), 1)
		a.Equal(len(ps.errors), 2)
	})

	resp, err := http.Get(srv.URL + "/params/int/1/2/str")
	a.NotError(err).NotNil(resp)
}

func TestParams_Bool_MustBool(t *testing.T) {
	a := assert.New(t)
	mux := mux.New(true, true, nil, nil)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.GetFunc("/params/bool/{b1}/{b2}/{str}", func(w http.ResponseWriter, r *http.Request) {
		ctx := newContextWithParam(w, r, a)
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

	resp, err := http.Get(srv.URL + "/params/bool/true/false/str")
	a.NotError(err).NotNil(resp)
}

func TestParams_String_MustString(t *testing.T) {
	a := assert.New(t)
	mux := mux.New(true, true, nil, nil)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.GetFunc("/params/string/{s1}/{s2}", func(w http.ResponseWriter, r *http.Request) {
		ctx := newContextWithParam(w, r, a)
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

	resp, err := http.Get(srv.URL + "/params/string/str1/str2")
	a.NotError(err).NotNil(resp)
}

func TestParams_Float_MustFloat(t *testing.T) {
	a := assert.New(t)
	mux := mux.New(true, true, nil, nil)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.GetFunc("/params/float/{f1}/{f2}/{str}", func(w http.ResponseWriter, r *http.Request) {
		ctx := newContextWithParam(w, r, a)
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

	resp, err := http.Get(srv.URL + "/params/float/1.1/2.2/str")
	a.NotError(err).NotNil(resp)
}

func TestParams_ParamID(t *testing.T) {
	a := assert.New(t)
	mux := mux.New(true, true, nil, nil)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.GetFunc("/params/paramid/{i1}/{i2}/{str}", func(w http.ResponseWriter, r *http.Request) {
		ctx := newContextWithParam(w, r, a)

		i1, err := ctx.ParamID("i1")
		a.NotError(err).Equal(i1, 1)

		i2, err := ctx.ParamID("i2")
		a.Error(err).Equal(i2, 0)

		i3, err := ctx.ParamID("i3")
		a.Error(err).Equal(i3, 0)
	})

	resp, err := http.Get(srv.URL + "/params/paramid/1/-2/str")
	a.NotError(err).NotNil(resp)
}

func TestParams_ParamInt64(t *testing.T) {
	a := assert.New(t)
	mux := mux.New(true, true, nil, nil)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.GetFunc("/params/paramint64/{i1}/{i2}/{str}", func(w http.ResponseWriter, r *http.Request) {
		ctx := newContextWithParam(w, r, a)

		i1, err := ctx.ParamInt64("i1")
		a.NotError(err).Equal(i1, 1)

		i2, err := ctx.ParamInt64("i2")
		a.NotError(err).Equal(i2, -2)

		i3, err := ctx.ParamInt64("i3")
		a.Error(err).Equal(i3, 0)
	})

	resp, err := http.Get(srv.URL + "/params/paramint64/1/-2/str")
	a.NotError(err).NotNil(resp)
}
