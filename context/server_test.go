// SPDX-License-Identifier: MIT

package context

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
	"github.com/issue9/logs/v2"

	"github.com/issue9/web/context/mimetype"
	"github.com/issue9/web/context/mimetype/gob"
	"github.com/issue9/web/context/mimetype/mimetypetest"
)

var f201 = func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusCreated)
	_, err := w.Write([]byte("1234567890"))
	if err != nil {
		println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// 声明一个 server 实例
func newServer(a *assert.Assertion) *Server {
	srv := newEmptyServer(a)

	err := srv.AddMarshals(map[string]mimetype.MarshalFunc{
		"application/json":       json.Marshal,
		"application/xml":        xml.Marshal,
		mimetype.DefaultMimetype: gob.Marshal,
		mimetypetest.Mimetype:    mimetypetest.TextMarshal,
	})
	a.NotError(err)

	err = srv.AddUnmarshals(map[string]mimetype.UnmarshalFunc{
		"application/json":       json.Unmarshal,
		"application/xml":        xml.Unmarshal,
		mimetype.DefaultMimetype: gob.Unmarshal,
		mimetypetest.Mimetype:    mimetypetest.TextUnmarshal,
	})
	a.NotError(err)

	return srv
}

func newEmptyServer(a *assert.Assertion) *Server {
	srv, err := NewServer(logs.New(), nil, false, false, "")
	a.NotError(err).NotNil(srv)
	return srv
}

func TestServer_AddStatic(t *testing.T) {
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

	// static 中定义的静态文件
	buf.Reset()
	srv.Get("/client/file1.txt").
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

	// 不存在的文件，测试 internal/fileserver 是否启作用
	srv.Get("/client/dir/not-exists.txt").
		Do().
		Status(http.StatusNotFound).
		StringBody("error handler test")
}

func TestServer_SetDebugger(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	srv := rest.NewServer(t, server.Handler(), nil)
	defer srv.Close()

	srv.Get("/debug/pprof/").Do().Status(http.StatusNotFound)
	srv.Get("/debug/vars").Do().Status(http.StatusNotFound)
	server.SetDebugger("/debug/pprof/", "/vars")
	srv.Get("/debug/pprof/").Do().Status(http.StatusOK)
	srv.Get("/vars").Do().Status(http.StatusOK)
}

func TestServer_URL_Path(t *testing.T) {
	a := assert.New(t)

	data := []*struct {
		root, input, url, path string
	}{
		{},

		{
			root:  "",
			input: "/abc",
			url:   "/abc",
			path:  "/abc",
		},

		{
			root:  "/",
			input: "/abc/def",
			url:   "/abc/def",
			path:  "/abc/def",
		},

		{
			root:  "https://localhost/",
			input: "/abc/def",
			url:   "https://localhost/abc/def",
			path:  "/abc/def",
		},
		{
			root:  "https://localhost",
			input: "/abc/def",
			url:   "https://localhost/abc/def",
			path:  "/abc/def",
		},
		{
			root:  "https://localhost",
			input: "abc/def",
			url:   "https://localhost/abc/def",
			path:  "/abc/def",
		},

		{
			root:  "https://localhost/",
			input: "",
			url:   "https://localhost",
			path:  "",
		},

		{
			root:  "https://example.com:8080/def/",
			input: "",
			url:   "https://example.com:8080/def",
			path:  "/def",
		},

		{
			root:  "https://example.com:8080/def/",
			input: "abc",
			url:   "https://example.com:8080/def/abc",
			path:  "/def/abc",
		},
	}

	for i, item := range data {
		srv, err := NewServer(logs.New(), DefaultResultBuilder, false, false, item.root)
		a.NotError(err, "error %s at %d", err, i).
			NotNil(srv)

		a.Equal(srv.URL(item.input), item.url, "not equal @%d,v1=%s,v2=%s", i, srv.URL(item.input), item.url)
		a.Equal(srv.Path(item.input), item.path, "not equal @%d,v1=%s,v2=%s", i, srv.Path(item.input), item.path)
	}
}
