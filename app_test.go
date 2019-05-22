// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"

	"github.com/issue9/web/app"
	"github.com/issue9/web/internal/resulttest"
	"github.com/issue9/web/internal/webconfig"
	"github.com/issue9/web/mimetype"
	"github.com/issue9/web/mimetype/gob"
	"github.com/issue9/web/mimetype/mimetypetest"
)

func getResult(status, code int, message string) app.Result {
	return resulttest.New(status, code, message)
}

func initApp(a *assert.Assertion) {
	defaultApp = nil
	a.NotError(Classic("./testdata", getResult))
	a.NotNil(defaultApp)
	a.Equal(defaultApp, App())

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
}

func TestClassic(t *testing.T) {
	a := assert.New(t)
	initApp(a)

	a.Panic(func() {
		a.NotError(Classic("./testdata", getResult))
	})

	a.True(IsDebug())
	a.Equal(URL("/test/abc.png"), "http://localhost:8082/test/abc.png")
	a.Equal(Path("/test/abc.png"), "/test/abc.png")
}

func TestTime(t *testing.T) {
	a := assert.New(t)
	initApp(a)

	a.Equal(Location(), time.UTC)
	a.Equal(Now().Location(), Location())
	a.Equal(Now().Unix(), time.Now().Unix())
}

func TestModules(t *testing.T) {
	a := assert.New(t)
	initApp(a)
	exit := make(chan bool, 1)

	m1 := NewModule("m1", "m1 desc", "m2")
	m1.AddInit(func() error {
		println("m1")
		return nil
	}, "init")
	m1.PostFunc("/post/m1", func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r)
		ctx.Render(http.StatusCreated, "m1", nil)
	})

	m2 := NewModule("m2", "m2 desc")
	a.NotNil(m2)

	a.NotError(InitModules(""))
	a.Equal(2, len(Modules())) //  m1,m2

	go func() {
		err := Serve()
		if err != http.ErrServerClosed {
			a.NotError(err)
		}
		exit <- true
	}()
	time.Sleep(500 * time.Microsecond)

	rest.NewRequest(a, nil, http.MethodPost, URL("/post")).
		Header("Accept", "application/json").
		Do().
		Status(http.StatusNotFound)

	rest.NewRequest(a, nil, http.MethodPost, URL("/post/m1")).
		Header("Accept", "application/json").
		Do().
		Status(http.StatusCreated).
		StringBody("\"m1\"") // json 格式的字符串

	ctx, c := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer c()
	a.NotError(Shutdown(ctx))

	<-exit
}

func TestFile(t *testing.T) {
	a := assert.New(t)
	initApp(a)

	path, err := filepath.Abs("./testdata/web.yaml")
	a.NotError(err)

	a.Equal(File("web.yaml"), path)

	conf1 := &webconfig.WebConfig{}
	a.NotError(LoadFile("web.yaml", conf1))
	a.NotNil(conf1)

	file, err := os.Open(path)
	a.NotError(err).NotNil(file)
	conf2 := &webconfig.WebConfig{}
	a.NotError(Load(file, ".yaml", conf2))
	a.NotNil(conf2)
	a.Equal(conf1, conf2)
}

func TestNewContext(t *testing.T) {
	a := assert.New(t)
	initApp(a)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept", "application/json")
	ctx := NewContext(w, r)
	a.NotNil(ctx).
		Equal(ctx.Response, w).
		Equal(ctx.Request, r).
		Equal(ctx.OutputCharsetName, "utf-8").
		Equal(ctx.OutputMimeTypeName, "application/json")
}

func TestGrace(t *testing.T) {
	if runtime.GOOS == "windows" { // windows 不支持 os.Process.Signal
		return
	}

	a := assert.New(t)
	exit := make(chan bool, 1)
	initApp(a)

	Grace(300*time.Millisecond, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		Serve()
		exit <- true
	}()
	time.Sleep(300 * time.Microsecond)

	p, err := os.FindProcess(os.Getpid())
	a.NotError(err).NotNil(p)
	a.NotError(p.Signal(syscall.SIGTERM))

	<-exit
}
