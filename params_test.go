// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestParams_empty(t *testing.T) {
	a := assert.New(t)
	app, err := NewApp("./testdata", nil)
	a.NotError(err).NotNil(app)

	srv := httptest.NewServer(app.Router().Mux())
	defer srv.Close()

	app.Router().GetFunc("/params/empty", func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r, nil)
		ps := ctx.Params()

		a.Equal(ps.Int("id1"), 0)
		a.Equal(ps.MustInt("id2", 2), 2)
	})

	resp, err := http.Get(srv.URL + "/params/empty")
	a.NotError(err).NotNil(resp)
}

func TestParams_Int_MustInt(t *testing.T) {
	a := assert.New(t)
	app, err := NewApp("./testdata", nil)
	a.NotError(err).NotNil(app)

	srv := httptest.NewServer(app.Router().Mux())
	defer srv.Close()

	app.Router().GetFunc("/params/int/{i1:\\d+}/{i2:\\d+}/{str}", func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r, nil)
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
	})

	resp, err := http.Get(srv.URL + "/params/int/1/2/str")
	a.NotError(err).NotNil(resp)
}

func TestParams_Bool_MustBool(t *testing.T) {
	a := assert.New(t)
	app, err := NewApp("./testdata", nil)
	a.NotError(err).NotNil(app)

	srv := httptest.NewServer(app.Router().Mux())
	defer srv.Close()

	app.Router().GetFunc("/params/bool/{b1}/{b2}/{str}", func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r, nil)
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
		a.True(ps.MustBool("str", true), true)
		a.Equal(len(ps.errors), 2)
	})

	resp, err := http.Get(srv.URL + "/params/bool/true/false/str")
	a.NotError(err).NotNil(resp)
}

func TestParams_String_MustString(t *testing.T) {
	a := assert.New(t)
	app, err := NewApp("./testdata", nil)
	a.NotError(err).NotNil(app)

	srv := httptest.NewServer(app.Router().Mux())
	defer srv.Close()

	app.Router().GetFunc("/params/string/{s1}/{s2}", func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r, nil)
		ps := ctx.Params()

		a.Equal(ps.String("s1"), "str1")
		a.Equal(ps.String("s2"), "str2")
		a.Nil(ps.Result(40001))

		// 不存在的参数，添加错误信息
		a.Equal(ps.String("s3"), "")
		a.Equal(len(ps.errors), 1)

		// MustString() 不会增加错误信息
		a.Equal(ps.MustString("s3", "str3"), "str3")
		a.Equal(len(ps.errors), 1)
		a.NotNil(ps.Result(40001))
	})

	resp, err := http.Get(srv.URL + "/params/string/str1/str2")
	a.NotError(err).NotNil(resp)
}

func TestParams_Float_MustFloat(t *testing.T) {
	a := assert.New(t)
	app, err := NewApp("./testdata", nil)
	a.NotError(err).NotNil(app)

	srv := httptest.NewServer(app.Router().Mux())
	defer srv.Close()

	app.Router().GetFunc("/params/float/{f1}/{f2}/{str}", func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r, nil)
		ps := ctx.Params()

		a.Equal(ps.Float64("f1"), 1.1)
		a.Equal(ps.Float64("f2"), 2.2)
		a.True(ps.OK(400001))

		// 不存在的参数，添加错误信息
		a.Equal(ps.Float64("f3"), 0.0)
		a.Equal(len(ps.errors), 1)

		// MustFloat64() 不会增加错误信息
		a.Equal(ps.MustFloat64("id3", 3.3), 3.3)
		a.Equal(len(ps.errors), 1)

		// MustFloat64() 无法转换，会返回默认值，且添加错误信息
		a.Equal(ps.MustFloat64("str", 3.3), 3.3)
		a.Equal(len(ps.errors), 2)

		a.False(ps.OK(400001))
	})

	resp, err := http.Get(srv.URL + "/params/float/1.1/2.2/str")
	a.NotError(err).NotNil(resp)
}

func TestParams_ParamID(t *testing.T) {
	a := assert.New(t)
	app, err := NewApp("./testdata", nil)
	a.NotError(err).NotNil(app)

	srv := httptest.NewServer(app.Router().Mux())
	defer srv.Close()

	app.Router().GetFunc("/params/paramid/{i1}/{i2}/{str}", func(w http.ResponseWriter, r *http.Request) {
		ctx := app.NewContext(w, r, nil)

		i1, ok := ctx.ParamID("i1", 40001)
		a.True(ok).Equal(i1, 1)

		i2, ok := ctx.ParamID("i2", 40001)
		a.False(ok).Equal(i2, 0)

		i3, ok := ctx.ParamID("i3", 40001)
		a.False(ok).Equal(i3, 0)
	})

	resp, err := http.Get(srv.URL + "/params/paramid/1/-2/str")
	a.NotError(err).NotNil(resp)
}

func TestParams_ParamInt64(t *testing.T) {
	a := assert.New(t)
	app, err := NewApp("./testdata", nil)
	a.NotError(err).NotNil(app)

	srv := httptest.NewServer(app.Router().Mux())
	defer srv.Close()

	app.Router().GetFunc("/params/paramint64/{i1}/{i2}/{str}", func(w http.ResponseWriter, r *http.Request) {
		ctx := app.NewContext(w, r, nil)

		i1, ok := ctx.ParamInt64("i1", 40001)
		a.True(ok).Equal(i1, 1)

		i2, ok := ctx.ParamInt64("i2", 40001)
		a.True(ok).Equal(i2, -2)

		i3, ok := ctx.ParamInt64("i3", 40001)
		a.False(ok).Equal(i3, 0)
	})

	resp, err := http.Get(srv.URL + "/params/paramint64/1/-2/str")
	a.NotError(err).NotNil(resp)
}
