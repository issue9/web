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
	"github.com/issue9/middleware/compress"
	"github.com/issue9/mux/v2"

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
	b := newEmptyServer(a)

	err := b.AddMarshals(map[string]mimetype.MarshalFunc{
		"application/json":       json.Marshal,
		"application/xml":        xml.Marshal,
		mimetype.DefaultMimetype: gob.Marshal,
		mimetypetest.Mimetype:    mimetypetest.TextMarshal,
	})
	a.NotError(err)

	err = b.AddUnmarshals(map[string]mimetype.UnmarshalFunc{
		"application/json":       json.Unmarshal,
		"application/xml":        xml.Unmarshal,
		mimetype.DefaultMimetype: gob.Unmarshal,
		mimetypetest.Mimetype:    mimetypetest.TextUnmarshal,
	})
	a.NotError(err)

	b.Compresses = map[string]compress.WriterFunc{
		"gzip":    compress.NewGzip,
		"deflate": compress.NewDeflate,
	}

	return b
}

func newEmptyServer(a *assert.Assertion) *Server {
	p := mux.New(false, false, false, nil, nil).Prefix("")
	a.NotNil(p)

	return NewServer(logs.New(), p, DefaultResultBuilder)
}

func TestMiddlewares(t *testing.T) {
	a := assert.New(t)
	app := newServer(a)
	err := app.errorHandlers.Add(func(w http.ResponseWriter, status int) {
		w.WriteHeader(status)
		_, err := w.Write([]byte("error handler test"))
		a.NotError(err)
	}, http.StatusNotFound)
	a.NotError(err)

	app.Router().Mux().GetFunc("/m1/test", f201)
	app.Static = map[string]string{
		"/client": "./testdata/",
	}
	app.errorHandlers.Add(func(w http.ResponseWriter, status int) {
		w.WriteHeader(status)
		_, err := w.Write([]byte("error handler test"))
		a.NotError(err)
	}, http.StatusNotFound)

	srv := rest.NewServer(t, app.Handler(), nil)
	defer srv.Close()

	buf := new(bytes.Buffer)
	srv.NewRequest(http.MethodGet, "/m1/test").
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
	srv.NewRequest(http.MethodGet, "/not-exists.txt").
		Do().
		Status(http.StatusNotFound).
		StringBody("error handler test")

	// static 中定义的静态文件
	buf.Reset()
	srv.NewRequest(http.MethodGet, "/client/file1.txt").
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
	srv.NewRequest(http.MethodGet, "/client/dir/not-exists.txt").
		Do().
		Status(http.StatusNotFound).
		StringBody("error handler test")
}

func TestBuiler_buildDebug(t *testing.T) {
	a := assert.New(t)

	b := newServer(a)
	srv := rest.NewServer(t, b.buildDebug(http.HandlerFunc(f201)), nil)
	defer srv.Close()

	// Debug = false
	srv.NewRequest(http.MethodGet, "/debug/pprof/").
		Do().
		Status(http.StatusCreated)

	// 命中 /debug/pprof/cmdline
	b.Debug = true
	srv.NewRequest(http.MethodGet, "/debug/pprof/").
		Do().
		Status(http.StatusOK)

	srv.NewRequest(http.MethodGet, "/debug/pprof/cmdline").
		Do().
		Status(http.StatusOK)

	srv.NewRequest(http.MethodGet, "/debug/pprof/trace").
		Do().
		Status(http.StatusOK)

	srv.NewRequest(http.MethodGet, "/debug/pprof/symbol").
		Do().
		Status(http.StatusOK)

	// /debug/vars
	srv.NewRequest(http.MethodGet, "/debug/vars").
		Do().
		Status(http.StatusOK)

	// 命中 h201
	srv.NewRequest(http.MethodGet, "/debug/").
		Do().
		Status(http.StatusCreated)
}
