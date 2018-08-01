// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
	"github.com/issue9/web/encoding"
	"github.com/issue9/web/encoding/encodingtest"
	"github.com/issue9/web/encoding/gob"
)

var testdata = ""

func TestMain(m *testing.M) {
	if err := Init("./internal/app/testdata/"); err != nil {
		panic(err)
	}

	if err := encoding.AddMarshal("text/plain", encodingtest.TextMarshal); err != nil {
		panic(err)
	}

	if err := encoding.AddMarshal(encoding.DefaultMimeType, gob.Marshal); err != nil {
		panic(err)
	}

	if err := encoding.AddUnmarshal(encoding.DefaultMimeType, gob.Unmarshal); err != nil {
		panic(err)
	}

	// m1 的路由项依赖 m2 的初始化数据
	m1 := NewModule("m1", "m1 desc", "m2")
	if m1 == nil {
		panic("m1==nil")
	}
	m1.AddInit(func() error {
		m1.PostFunc("/post/"+testdata, func(w http.ResponseWriter, r *http.Request) {
			if testdata != "m2" {
				panic("testdata!=m2")
			}
			ctx := NewContext(w, r)
			ctx.Render(http.StatusCreated, testdata, nil)
		})
		return nil
	})

	m2 := NewModule("m2", "m2 desc")
	if m2 == nil {
		panic("m2==nil")
	}
	m2.AddInit(func() error {
		testdata = "m2"
		return nil
	})
}

func TestInit(t *testing.T) {
	a := assert.New(t)
	a.Error(Init("./internal/app/testdata/"))
}

func TestIsDebug(t *testing.T) {
	a := assert.New(t)
	a.True(IsDebug())
}

func TestFile(t *testing.T) {
	a := assert.New(t)
	path, err := filepath.Abs("./internal/app/testdata/abc.yaml")
	a.NotError(err)
	a.Equal(File("/abc.yaml"), path)
}

func TestNewContext(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept", encoding.DefaultMimeType)
	ctx := NewContext(w, r)
	a.NotNil(ctx).
		Equal(ctx.Response, w).
		Equal(ctx.Request, r).
		Equal(ctx.OutputCharsetName, "utf-8").
		Equal(ctx.OutputMimeTypeName, encoding.DefaultMimeType)
}

func TestModules(t *testing.T) {
	a := assert.New(t)
	ms := Modules()
	a.Equal(2, len(ms)) // 由 testmain 中初始化的 m1,m2
}

func TestURL(t *testing.T) {
	a := assert.New(t)
	a.Equal(URL("/test/abc.png"), "http://localhost:8082/test/abc.png")
}

func TestHandler(t *testing.T) {
	a := assert.New(t)

	h, err := Handler()
	a.NotError(err).NotNil(h)

	srv := rest.NewServer(t, h, nil)
	a.NotNil(srv)
	defer srv.Close()

	srv.NewRequest(http.MethodPost, "/post").
		Header("Accept", "text/plain").
		Do().
		Status(http.StatusNotFound)

	srv.NewRequest(http.MethodPost, "/post/m2").
		Header("Accept", "text/plain").
		Do().
		Status(http.StatusCreated).
		Body([]byte(testdata))
}
