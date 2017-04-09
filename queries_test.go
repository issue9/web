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

func TestQueries_Int(t *testing.T) {
	a := assert.New(t)
	app, err := NewApp("./testdata")
	a.NotError(err).NotNil(app)

	srv := httptest.NewServer(app.Router().Mux())
	defer srv.Close()

	app.Router().GetFunc("/queries/int", func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r, nil)
		q := ctx.Queries()

		a.Equal(q.Int("i1", 9), 1)
		a.Equal(q.Int("i2", 9), 2)
		a.Equal(q.Int("i3", 9), 9)
		a.True(q.OK(40001))

		// 无法转换，会返回默认值，且添加错误信息
		a.Equal(q.Int("str", 3), 3)
		a.Equal(len(q.errors), 1)
		a.False(q.OK(40001))
	})

	resp, err := http.Get(srv.URL + "/queries/int?i1=1&i2=2&str=str")
	a.NotError(err).NotNil(resp)
}

func TestQueries_Int64(t *testing.T) {
	a := assert.New(t)
	app, err := NewApp("./testdata")
	a.NotError(err).NotNil(app)

	srv := httptest.NewServer(app.Router().Mux())
	defer srv.Close()

	app.Router().GetFunc("/queries/int64", func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r, nil)
		q := ctx.Queries()

		a.Equal(q.Int64("i1", 9), 1)
		a.Equal(q.Int64("i2", 9), 2)
		a.Equal(q.Int64("i3", 9), 9)
		a.True(q.OK(40001))

		// 无法转换，会返回默认值，且添加错误信息
		a.Equal(q.Int64("str", 3), 3)
		a.Equal(len(q.errors), 1)
		a.False(q.OK(40001))
	})

	resp, err := http.Get(srv.URL + "/queries/int64?i1=1&i2=2&str=str")
	a.NotError(err).NotNil(resp)
}

func TestQueries_String(t *testing.T) {
	a := assert.New(t)
	app, err := NewApp("./testdata")
	a.NotError(err).NotNil(app)

	srv := httptest.NewServer(app.Router().Mux())
	defer srv.Close()

	app.Router().GetFunc("/queries/string", func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r, nil)
		q := ctx.Queries()

		a.Equal(q.String("s1", "9"), "1")
		a.Equal(q.String("s2", "9"), "2")
		a.Equal(q.String("s3", "9"), "9")
		a.True(q.OK(40001))
	})

	resp, err := http.Get(srv.URL + "/queries/string?s1=1&s2=2")
	a.NotError(err).NotNil(resp)
}

func TestQueries_Bool(t *testing.T) {
	a := assert.New(t)
	app, err := NewApp("./testdata")
	a.NotError(err).NotNil(app)

	srv := httptest.NewServer(app.Router().Mux())
	defer srv.Close()

	app.Router().GetFunc("/queries/bool", func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r, nil)
		q := ctx.Queries()

		a.True(q.Bool("b1", false))
		a.True(q.Bool("b2", false))
		a.False(q.Bool("b3", false))
		a.True(q.OK(40001)).Nil(q.Result(400001))

		// 无法转换，会返回默认值，且添加错误信息
		a.False(q.Bool("str", false))
		a.Equal(len(q.errors), 1)
		a.False(q.OK(40001)).NotNil(q.Result(400001))
	})

	resp, err := http.Get(srv.URL + "/queries/bool?b1=true&b2=true&str=str")
	a.NotError(err).NotNil(resp)
}

func TestQueries_Float64(t *testing.T) {
	a := assert.New(t)
	app, err := NewApp("./testdata")
	a.NotError(err).NotNil(app)

	srv := httptest.NewServer(app.Router().Mux())
	defer srv.Close()

	app.Router().GetFunc("/queries/float64", func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r, nil)
		q := ctx.Queries()

		a.Equal(q.Float64("i1", 9.9), 1.1)
		a.Equal(q.Float64("i2", 9.9), 2)
		a.Equal(q.Float64("i3", 9.9), 9.9)
		a.True(q.OK(40001))

		// 无法转换，会返回默认值，且添加错误信息
		a.Equal(q.Float64("str", 3), 3)
		a.Equal(len(q.errors), 1)
		a.False(q.OK(40001))
	})

	resp, err := http.Get(srv.URL + "/queries/float64?i1=1.1&i2=2&str=str")
	a.NotError(err).NotNil(resp)
}
