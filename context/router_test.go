// SPDX-License-Identifier: MIT

package context

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
)

var f1 = func(ctx *Context) { ctx.Render(http.StatusOK, nil, nil) }

func TestPrefix(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	srv := rest.NewServer(t, server.Handler(), nil)

	p := server.Prefix("/p")
	a.NotNil(p)

	path := "/path"
	a.NotError(p.Handle(path, f1, http.MethodGet, http.MethodDelete))
	srv.Get("/root/p" + path).Do().Status(http.StatusOK)
	srv.Delete("/root/p" + path).Do().Status(http.StatusOK)
	srv.Post("/root/p"+path, nil).Do().Status(http.StatusMethodNotAllowed)

	p.Post(path, f1)
	srv.Post("/root/p"+path, nil).Do().Status(http.StatusOK)

	p.Patch(path, f1)
	srv.Patch("/root/p"+path, nil).Do().Status(http.StatusOK)

	p.Options(path, "abc")
	srv.NewRequest(http.MethodOptions, "/root/p"+path).
		Do().
		Status(http.StatusOK).
		Header("allow", "abc")

	p.Remove(path, http.MethodDelete)
	srv.Delete("/root/p" + path).Do().Status(http.StatusMethodNotAllowed)

	// resource

	path = "/resources/{id}"
	res := p.Resource(path)
	res.Get(f1).Delete(f1)
	srv.Get("/root/p" + path).Do().Status(http.StatusOK)
	srv.Delete("/root/p" + path).Do().Status(http.StatusOK)

	res.Remove(http.MethodDelete)
	srv.Delete("/root/p" + path).Do().Status(http.StatusMethodNotAllowed)
	res.Remove(http.MethodGet)
	srv.Delete("/root/p" + path).Do().Status(http.StatusNotFound)

	// prefix 带域名
	p.Delete("example.com/domain", f1)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/root/p/domain", nil)
	server.Handler().ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusNotFound)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodDelete, "http://example.com:88/domain", nil)
	server.Handler().ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusOK)
}

func TestResource(t *testing.T) {
	a := assert.New(t)

	server := newServer(a)

	path := "/path"
	res := server.Resource(path)
	a.NotNil(res)

	srv := rest.NewServer(t, server.Handler(), nil)

	res.Get(f1)
	srv.Get("/root" + path).Do().Status(http.StatusOK)

	res.Delete(f1)
	srv.Delete("/root" + path).Do().Status(http.StatusOK)

	res.Patch(f1)
	srv.Patch("/root"+path, nil).Do().Status(http.StatusOK)

	res.Put(f1)
	srv.Put("/root"+path, nil).Do().Status(http.StatusOK)

	res.Post(f1)
	srv.Post("/root"+path, nil).Do().Status(http.StatusOK)

	res.Remove(http.MethodPost)
	srv.Post("/root"+path, nil).Do().Status(http.StatusMethodNotAllowed)

	res.Options("def")
	srv.NewRequest(http.MethodOptions, "/root"+path).Do().Header("allow", "def")

	// 带域名
	server = newServer(a)
	res = server.Resource("example.com/resource")
	res.Delete(f1)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/root/resource", nil)
	server.Handler().ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusNotFound)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodDelete, "http://example.com:88/resource", nil)
	server.Handler().ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusOK)
}

func TestServer_Handle(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	srv := rest.NewServer(t, server.Handler(), nil)

	path := "/path"
	a.NotError(server.Handle(path, f1, http.MethodGet, http.MethodDelete))
	srv.Get("/root" + path).Do().Status(http.StatusOK)
	srv.Delete("/root" + path).Do().Status(http.StatusOK)
	srv.Post("/root"+path, nil).Do().Status(http.StatusMethodNotAllowed)

	// 不指定请求方法，表示所有请求方法
	path = "/path1"
	a.NotError(server.Handle(path, f1))
	srv.Delete("/root" + path).Do().Status(http.StatusOK)
	srv.Patch("/root"+path, nil).Do().Status(http.StatusOK)

	path = "/path2"

	srv.Delete("/root" + path).Do().Status(http.StatusNotFound)

	server.Delete(path, f1)
	srv.Delete("/root" + path).Do().Status(http.StatusOK)

	server.Get(path, f1)
	srv.Get("/root" + path).Do().Status(http.StatusOK)

	server.Post(path, f1)
	srv.Post("/root"+path, nil).Do().Status(http.StatusOK)

	server.Patch(path, f1)
	srv.Patch("/root"+path, nil).Do().Status(http.StatusOK)

	server.Put(path, f1)
	srv.Put("/root"+path, nil).Do().Status(http.StatusOK)

	srv.NewRequest(http.MethodOptions, "/root"+path).
		Do().
		Status(http.StatusOK).
		Header("Allow", "DELETE, GET, HEAD, OPTIONS, PATCH, POST, PUT")

	// 自定义 options
	server.Options(path, "abc")
	srv.NewRequest(http.MethodOptions, "/root"+path).
		Do().
		Status(http.StatusOK).
		Header("Allow", "abc")

	server.Remove(path, http.MethodOptions)
	srv.NewRequest(http.MethodOptions, "/root"+path).
		Do().
		Status(http.StatusMethodNotAllowed)

	//  带域名
	server.Get("example.com/domain", f1)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/root/domain", nil)
	server.Handler().ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusNotFound)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://example.com:88/domain", nil)
	server.Handler().ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusOK)
}

func TestServer_Static(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	server.SetErrorHandle(func(w http.ResponseWriter, status int) {
		w.WriteHeader(status)
		_, err := w.Write([]byte("error handler test"))
		a.NotError(err)
	}, http.StatusNotFound)

	server.Router().Mux().GetFunc("/m1/test", f201)
	server.AddStatic("/client", "./testdata/")
	server.SetErrorHandle(func(w http.ResponseWriter, status int) {
		w.WriteHeader(status)
		_, err := w.Write([]byte("error handler test"))
		a.NotError(err)
	}, http.StatusNotFound)

	srv := rest.NewServer(t, server.Handler(), nil)
	defer srv.Close()

	buf := new(bytes.Buffer)
	srv.Get("/m1/test").
		Header("Accept-Encoding", "gzip,deflate;q=0.8").
		Do().
		Status(http.StatusCreated).
		ReadBody(buf).
		Header("Content-Type", "text/html").
		Header("Content-Encoding", "gzip").
		Header("Vary", "Content-Encoding")
	reader, err := gzip.NewReader(buf)
	a.NotError(err).NotNil(reader)
	data, err := ioutil.ReadAll(reader)
	a.NotError(err).NotNil(data)
	a.Equal(string(data), "1234567890")

	// not found
	// 返回 ErrorHandler 内容
	srv.Get("/not-exists.txt").
		Do().
		Status(http.StatusNotFound).
		StringBody("error handler test")

	// 定义的静态文件
	buf.Reset()
	srv.Get("/root/client/file1.txt").
		Header("Accept-Encoding", "gzip,deflate;q=0.8").
		Do().
		Status(http.StatusOK).
		ReadBody(buf).
		Header("Content-Type", "text/plain; charset=utf-8").
		Header("Content-Encoding", "gzip").
		Header("Vary", "Content-Encoding")
	reader, err = gzip.NewReader(buf)
	a.NotError(err).NotNil(reader)
	data, err = ioutil.ReadAll(reader)
	a.NotError(err).NotNil(data)
	a.Equal(string(data), "file1")

	// 删除
	server.RemoveStatic("/client")
	srv.Get("/root/client/file1.txt").
		Do().
		Status(http.StatusNotFound)

	// 带域名
	server.AddStatic("example.com/client", "./testdata/")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://example.com/client/file1.txt", nil)
	server.Handler().ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusOK)
}
