// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
)

func TestContext_ServeFile(t *testing.T) {
	a := assert.New(t)
	exit := make(chan bool, 1)

	s := newServer(a)
	defer func() {
		a.NotError(s.Close(0))
		<-exit
	}()

	a.NotPanic(func() {
		s.Router().Get("/path", func(ctx *Context) {
			ctx.ServeFile("./testdata/file1.txt", "index.html", map[string]string{"Test": "Test"})
		})

		s.Router().Get("/index", func(ctx *Context) {
			ctx.ServeFile("./testdata", "file1.txt", map[string]string{"Test": "Test"})
		})

		s.Router().Get("/not-exists", func(ctx *Context) {
			// file1.text 不存在
			ctx.ServeFile("./testdata/file1.text", "index.html", map[string]string{"Test": "Test"})
		})
	})

	go func() {
		a.Equal(s.Serve(), http.ErrServerClosed)
		exit <- true
	}()
	time.Sleep(500 * time.Millisecond)

	testDownload(a, "http://localhost:8080/root/path", http.StatusOK)
	testDownload(a, "http://localhost:8080/root/index", http.StatusOK)
	testDownloadNotFound(a, "http://localhost:8080/root/not-exits")
}

func TestContext_ServeFileFS(t *testing.T) {
	a := assert.New(t)
	exit := make(chan bool, 1)

	s := newServer(a)
	defer func() {
		a.NotError(s.Close(0))
		<-exit
	}()

	fs := os.DirFS("./testdata")

	s.Router().Get("/path", func(ctx *Context) {
		ctx.ServeFileFS(fs, "file1.txt", "index.html", map[string]string{"Test": "Test"})
	})

	s.Router().Get("/index", func(ctx *Context) {
		ctx.ServeFileFS(fs, ".", "file1.txt", map[string]string{"Test": "Test"})
	})

	s.Router().Get("/not-exists", func(ctx *Context) {
		// file.text 不存在
		ctx.ServeFileFS(fs, "file1.text", "index.html", map[string]string{"Test": "Test"})
	})

	go func() {
		a.Equal(s.Serve(), http.ErrServerClosed)
		exit <- true
	}()
	time.Sleep(500 * time.Millisecond)

	testDownload(a, "http://localhost:8080/root/path", http.StatusOK)
	testDownload(a, "http://localhost:8080/root/index", http.StatusOK)
	testDownloadNotFound(a, "http://localhost:8080/root/not-exits")
}

func TestContext_ServeContent(t *testing.T) {
	a := assert.New(t)
	exit := make(chan bool, 1)

	s := newServer(a)
	defer func() {
		a.NotError(s.Close(0))
		<-exit
	}()

	buf, err := ioutil.ReadFile("./testdata/file1.txt")
	a.NotError(err).NotNil(buf)

	s.Router().Get("/path", func(ctx *Context) {
		ctx.ServeContent(bytes.NewReader(buf), "name", time.Now(), map[string]string{"Test": "Test"})
	})

	go func() {
		a.Equal(s.Serve(), http.ErrServerClosed)
		exit <- true
	}()
	time.Sleep(500 * time.Millisecond)

	testDownload(a, "http://localhost:8080/root/path", http.StatusOK)
	testDownloadNotFound(a, "http://localhost:8080/root/path/not-exits")
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
