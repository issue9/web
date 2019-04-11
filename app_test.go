// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"

	"github.com/issue9/web/app"
	"github.com/issue9/web/internal/resulttest"
	"github.com/issue9/web/mimetype"
	"github.com/issue9/web/mimetype/gob"
	"github.com/issue9/web/mimetype/mimetypetest"
)

var testdata = ""

func getResult(status, code int, message string) app.Result {
	return resulttest.New(status, code, message)
}

func TestApp(t *testing.T) {
	a := assert.New(t)
	defaultApp = nil
	exit := make(chan bool, 1)

	a.NotError(Classic("./testdata/web.yaml", getResult))
	a.NotNil(defaultApp)
	a.Equal(defaultApp, App())

	a.Panic(func() {
		Classic("./testdata/web.yaml", getResult)
	})

	err := Mimetypes().AddMarshals(map[string]mimetype.MarshalFunc{
		"application/xml":        xml.Marshal,
		mimetype.DefaultMimetype: gob.Marshal,
		mimetypetest.MimeType:    mimetypetest.TextMarshal,
	})
	a.NotError(err)

	err = Mimetypes().AddUnmarshals(map[string]mimetype.UnmarshalFunc{
		"application/xml":        xml.Unmarshal,
		mimetype.DefaultMimetype: gob.Unmarshal,
		mimetypetest.MimeType:    mimetypetest.TextUnmarshal,
	})
	a.NotError(err)

	a.True(IsDebug())

	a.Equal(URL("/test/abc.png"), "http://localhost:8082/test/abc.png")

	testFile(a)

	// m1 的路由项依赖 m2 的初始化数据
	m1 := NewModule("m1", "m1 desc", "m2")
	m1.AddInit(func() error {
		if testdata != "m2" {
			panic("testdata!=m2")
		}

		Mux().PostFunc("/post/"+testdata, func(w http.ResponseWriter, r *http.Request) {
			ctx := NewContext(w, r)
			ctx.Render(http.StatusCreated, testdata, nil)
		})
		return nil
	}, "init")

	m2 := NewModule("m2", "m2 desc")
	m2.AddInit(func() error {
		testdata = "m2"
		return nil
	}, "init")

	a.NotError(InitModules(""))

	a.Equal(3, len(Modules())) //  m1,m2，以及自带的模块 web-core

	go func() {
		err := Serve()
		if err != http.ErrServerClosed {
			a.NotError(err)
		}
		exit <- true
	}()
	time.Sleep(500 * time.Microsecond)

	rest.NewRequest(a, nil, http.MethodPost, URL("/post")).
		Header("Accept", mimetype.DefaultMimetype).
		Do().
		Status(http.StatusNotFound)

	rest.NewRequest(a, nil, http.MethodPost, URL("/post/m2")).
		Header("Accept", mimetypetest.MimeType).
		Do().
		Status(http.StatusCreated).
		StringBody(testdata)

	a.NotError(Shutdown())

	<-exit
}

func testFile(a *assert.Assertion) {
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
