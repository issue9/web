// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
	"github.com/issue9/mux/v5/group"

	"github.com/issue9/web/content"
	"github.com/issue9/web/content/text"
	"github.com/issue9/web/content/text/testobject"
)

func TestContext_Vars(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("Accept", "*/*")
	w := httptest.NewRecorder()
	ctx := newServer(a).NewContext(w, r)

	type (
		t1 int
		t2 int64
		t3 = t2
	)
	var (
		v1 t1 = 1
		v2 t2 = 1
		v3 t3 = 1
	)

	ctx.Vars[v1] = 1
	ctx.Vars[v2] = 2
	ctx.Vars[v3] = 3

	a.Equal(ctx.Vars[v1], 1).Equal(ctx.Vars[v2], 3)
}

func TestServer_NewContext(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	srv := newServer(a)
	logwriter := new(bytes.Buffer)
	srv.Logs().DEBUG().SetOutput(logwriter)

	// 错误的 accept
	logwriter.Reset()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", "not")
	a.Panic(func() {
		srv.NewContext(w, r)
	})
	a.True(logwriter.Len() > 0)

	// 正常，未指定 Accept-Language 和 Accept-Charset 等不是必须的报头
	logwriter.Reset()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept", content.DefaultMimetype)
	a.NotPanic(func() {
		ctx := srv.NewContext(w, r)
		a.NotNil(ctx).
			Equal(logwriter.Len(), 0).
			Equal(ctx.OutputMimetypeName, content.DefaultMimetype)
	})
}

func TestContext_Read(t *testing.T) {
	a := assert.New(t)
	r := httptest.NewRequest(http.MethodPost, "/path", bytes.NewBufferString("test,123"))
	w := httptest.NewRecorder()
	r.Header.Set("Content-Type", text.Mimetype+"; charset=utf-8")
	ctx := newServer(a).NewContext(w, r)

	obj := &testobject.TextObject{}
	a.True(ctx.Read(obj, 41110))
	a.Equal(obj.Name, "test").Equal(obj.Age, 123)

	o := &struct{}{}
	a.False(ctx.Read(o, 41110))
	a.Equal(w.Result().StatusCode, http.StatusUnprocessableEntity)
}

func TestContext_ClientIP(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	ctx := newServer(a).NewContext(w, r)
	a.Equal(ctx.ClientIP(), r.RemoteAddr)

	// httptest.NewRequest 会直接将  remote-addr 赋值为 192.0.2.1 无法测试
	r, err := http.NewRequest(http.MethodPost, "/path", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("x-real-ip", "192.168.1.1:8080")
	ctx = newServer(a).NewContext(w, r)
	a.Equal(ctx.ClientIP(), "192.168.1.1:8080")

	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111")
	ctx = newServer(a).NewContext(w, r)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")

	// 测试获取 IP 报头的优先级
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111")
	r.Header.Set("x-real-ip", "192.168.2.2")
	ctx = newServer(a).NewContext(w, r)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")

	// 测试获取 IP 报头的优先级
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set("Remote-Addr", "192.168.2.0")
	r.Header.Set("x-forwarded-for", "192.168.2.1:8080,192.168.2.2:111")
	r.Header.Set("x-real-ip", "192.168.2.2")
	ctx = newServer(a).NewContext(w, r)
	a.Equal(ctx.ClientIP(), "192.168.2.1:8080")
}

func TestContext_ServeFile(t *testing.T) {
	a := assert.New(t)
	exit := make(chan bool, 1)

	s := newServer(a)
	defer func() {
		a.NotError(s.Close(0))
		<-exit
	}()
	router, err := s.NewRouter("default", "http://localhost:8080/root", group.MatcherFunc(group.Any))
	a.NotError(err).NotNil(router)

	a.NotPanic(func() {
		router.Get("/path", func(ctx *Context) Responser {
			return ctx.ServeFile("./testdata/file1.txt", "index.html", map[string]string{"Test": "Test"})
		})

		router.Get("/index", func(ctx *Context) Responser {
			return ctx.ServeFile("./testdata", "file1.txt", map[string]string{"Test": "Test"})
		})

		router.Get("/not-exists", func(ctx *Context) Responser {
			// file1.text 不存在
			return ctx.ServeFile("./testdata/file1.text", "index.html", map[string]string{"Test": "Test"})
		})
	})

	go func() {
		a.Equal(s.Serve("default", true), http.ErrServerClosed)
		exit <- true
	}()
	time.Sleep(500 * time.Millisecond)

	testDownload(a, "http://localhost:8080/root/path", http.StatusOK)
	testDownload(a, "http://localhost:8080/root/index", http.StatusOK)
	testDownloadNotFound(a, "http://localhost:8080/root/not-exists")
}

func TestContext_ServeFile_windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		return
	}

	a := assert.New(t)
	exit := make(chan bool, 1)

	s := newServer(a)
	defer func() {
		a.NotError(s.Close(0))
		<-exit
	}()
	router, err := s.NewRouter("default", "http://localhost:8080/root", group.MatcherFunc(group.Any))
	a.NotError(err).NotNil(router)

	a.NotPanic(func() {
		router.Get("/path", func(ctx *Context) Responser {
			return ctx.ServeFile(".\\testdata\\file1.txt", "index.html", map[string]string{"Test": "Test"})
		})

		router.Get("/index", func(ctx *Context) Responser {
			return ctx.ServeFile(".\\testdata", "file1.txt", map[string]string{"Test": "Test"})
		})

		router.Get("/not-exists", func(ctx *Context) Responser {
			// file1.text 不存在
			return ctx.ServeFile("c:\\not-exists-dir\\file1.text", "index.html", map[string]string{"Test": "Test"})
		})
	})

	go func() {
		a.Equal(s.Serve("default", true), http.ErrServerClosed)
		exit <- true
	}()
	time.Sleep(500 * time.Millisecond)

	testDownload(a, "http://localhost:8080/root/path", http.StatusOK)
	testDownload(a, "http://localhost:8080/root/index", http.StatusOK)
	testDownloadNotFound(a, "http://localhost:8080/root/not-exists")
}

func testDownload(a *assert.Assertion, path string, status int) {
	rest.NewRequest(a, nil, http.MethodGet, path).Do().
		Status(status).
		BodyNotNil().
		Header("Test", "Test")
}

func testDownloadNotFound(a *assert.Assertion, path string) {
	rest.NewRequest(a, nil, http.MethodGet, path).Do().
		Status(http.StatusNotFound)
}

func TestContext_Now(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := srv.NewContext(w, r)

	now := ctx.Now()
	a.Equal(now.Location(), srv.Location())
	a.Equal(now.Location(), srv.Now().Location())
}
