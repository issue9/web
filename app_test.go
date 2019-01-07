// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"

	"github.com/issue9/web/mimetype"
	"github.com/issue9/web/mimetype/gob"
)

var testdata = ""

func TestMain(m *testing.M) {
	if err := Init("./testdata"); err != nil {
		panic(err)
	}

	if defaultApp == nil {
		panic("defaultApp == nil")
	}

	err := Mimetypes().AddMarshals(map[string]mimetype.MarshalFunc{
		"application/json":       json.Marshal,
		"application/xml":        xml.Marshal,
		mimetype.DefaultMimetype: gob.Marshal,
	})
	if err != nil {
		panic(err)
	}

	err = Mimetypes().AddUnmarshals(map[string]mimetype.UnmarshalFunc{
		"application/json":       json.Unmarshal,
		"application/xml":        xml.Unmarshal,
		mimetype.DefaultMimetype: gob.Unmarshal,
	})
	if err != nil {
		panic(err)
	}

	// m1 的路由项依赖 m2 的初始化数据
	m1 := NewModule("m1", "m1 desc", "m2")
	m1.AddInit(func() error {
		if testdata != "m2" {
			panic("testdata!=m2")
		}

		m1.PostFunc("/post/"+testdata, func(w http.ResponseWriter, r *http.Request) {
			ctx := NewContext(w, r)
			ctx.Render(http.StatusCreated, testdata, nil)
		})
		return nil
	})

	m2 := NewModule("m2", "m2 desc")
	m2.AddInit(func() error {
		testdata = "m2"
		return nil
	})

	os.Exit(m.Run())
}

func TestIsDebug(t *testing.T) {
	a := assert.New(t)
	a.True(IsDebug())
}

func TestFile(t *testing.T) {
	a := assert.New(t)
	path, err := filepath.Abs("./testdata/abc.yaml")
	a.NotError(err)
	a.Equal(File("/abc.yaml"), path)
}

func TestNewContext(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept", mimetype.DefaultMimetype)
	ctx := NewContext(w, r)
	a.NotNil(ctx).
		Equal(ctx.Response, w).
		Equal(ctx.Request, r).
		Equal(ctx.OutputCharsetName, "utf-8").
		Equal(ctx.OutputMimeTypeName, mimetype.DefaultMimetype)
}

func TestModules(t *testing.T) {
	a := assert.New(t)
	ms := Modules()
	a.Equal(3, len(ms)) // 由 testmain 中初始化的 m1,m2，以及自带的模块 web-core
}

func TestURL(t *testing.T) {
	a := assert.New(t)
	a.Equal(URL("/test/abc.png"), "http://localhost:8082/test/abc.png")
}

func TestMux(t *testing.T) {
	a := assert.New(t)

	srv := rest.NewServer(t, Mux(), nil)
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
