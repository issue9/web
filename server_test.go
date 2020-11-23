// SPDX-License-Identifier: MIT

package web

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
	"github.com/issue9/logs/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/web/content"
	"github.com/issue9/web/content/gob"
	"github.com/issue9/web/content/mimetypetest"
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

var f202 = func(ctx *Context) {
	ctx.Response.WriteHeader(http.StatusAccepted)
	_, err := ctx.Response.Write([]byte("1234567890"))
	if err != nil {
		println(err)
		ctx.Response.WriteHeader(http.StatusInternalServerError)
	}
}

// 声明一个 server 实例
func newServer(a *assert.Assertion) *Server {
	o := &Options{Root: "/root"}
	srv, err := NewServer(logs.New(), o)
	a.NotError(err).NotNil(srv)

	// srv.Catalog 默认指向 message.DefaultCatalog
	a.NotError(message.SetString(language.Und, "lang", "und"))
	a.NotError(message.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(message.SetString(language.TraditionalChinese, "lang", "hant"))

	err = srv.Mimetypes().AddMarshals(map[string]content.MarshalFunc{
		"application/json":      json.Marshal,
		"application/xml":       xml.Marshal,
		content.DefaultMimetype: gob.Marshal,
		mimetypetest.Mimetype:   mimetypetest.TextMarshal,
	})
	a.NotError(err)

	err = srv.Mimetypes().AddUnmarshals(map[string]content.UnmarshalFunc{
		"application/json":      json.Unmarshal,
		"application/xml":       xml.Unmarshal,
		content.DefaultMimetype: gob.Unmarshal,
		mimetypetest.Mimetype:   mimetypetest.TextUnmarshal,
	})
	a.NotError(err)

	srv.AddResultMessage(411, 41110, "41110")

	return srv
}

func TestNewServer(t *testing.T) {
	a := assert.New(t)
	l := logs.New()
	srv, err := NewServer(l, &Options{})
	a.NotError(err).NotNil(srv)
	a.False(srv.Uptime().IsZero())
	a.Equal(l, srv.Logs())
	a.NotNil(srv.Cache())
	a.Equal(srv.catalog, message.DefaultCatalog)
	a.Equal(srv.Location(), time.Local)
	a.Equal(srv.httpServer.Handler, srv.middlewares)
	a.NotNil(srv.httpServer.BaseContext)
}

func TestGetServer(t *testing.T) {
	a := assert.New(t)
	type key int
	var k key = 0

	srv, err := NewServer(logs.New(), &Options{})
	srv.mimetypes.AddMarshal(mimetypetest.Mimetype, mimetypetest.TextMarshal)
	a.NotError(err).NotNil(srv)
	var isRequested bool

	srv.Router().Mux().GetFunc("/path", func(w http.ResponseWriter, r *http.Request) {
		s1 := GetServer(r)
		a.NotNil(s1).Equal(s1, srv)

		v := r.Context().Value(k)
		a.Nil(v)

		ctx := NewContext(w, r)
		a.NotNil(ctx)

		isRequested = true
	})
	go func() {
		srv.Serve()
	}()
	time.Sleep(500 * time.Millisecond)
	rest.NewRequest(a, nil, http.MethodGet, "http://localhost/path").
		Header("Accept", mimetypetest.Mimetype).
		Do().
		Success("未正确返回状态码")
	a.NotError(srv.Close(0))
	a.True(isRequested, "未正常访问 /path")

	// 不是从 Server 生成的 *http.Request，则会 panic
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	a.Panic(func() {
		GetServer(r)
	})

	// BaseContext

	srv, err = NewServer(logs.New(), &Options{
		HTTPServer: func(s *http.Server) {
			s.BaseContext = func(n net.Listener) context.Context {
				return context.WithValue(context.Background(), k, 1)
			}
		},
	})
	a.NotError(err).NotNil(srv)

	isRequested = false
	srv.Router().Mux().GetFunc("/path", func(w http.ResponseWriter, r *http.Request) {
		s1 := GetServer(r)
		a.NotNil(s1).Equal(s1, srv)

		v := r.Context().Value(k)
		a.Equal(v, 1)

		isRequested = true
	})
	go func() {
		srv.Serve()
	}()
	time.Sleep(500 * time.Millisecond)
	rest.NewRequest(a, nil, http.MethodGet, "http://localhost/path").Do().Success()
	a.NotError(srv.Close(0))
	a.True(isRequested, "未正常访问 /path")
}

func TestServer_Vars(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)

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

	srv.Set(v1, 1)
	srv.Set(v2, 2)
	srv.Set(v3, 3)

	a.Equal(srv.Get(v1), 1).Equal(srv.Get(v2), 3)
}

func TestServer_Serve(t *testing.T) {
	a := assert.New(t)
	exit := make(chan bool, 1)

	server := newServer(a)
	server.Router().Get("/mux/test", f202)

	m1, err := server.NewModule("m1", "m1 desc")
	a.NotError(err).NotNil(m1)
	m1.Get("/m1/test", f202)
	m1.NewTag("tag1")

	m2, err := server.NewModule("m2", "m2 desc", "m1")
	a.NotError(err).NotNil(m2)
	m2.Get("/m2/test", func(ctx *Context) {
		srv := ctx.Server()
		a.NotNil(srv)
		a.Equal(2, len(srv.Modules()))
		a.Equal(1, len(srv.Tags())).
			Equal(srv.Tags()["m1"], []string{"tag1"}).
			Nil(srv.Tags()["m2"])

		ctx.Response.WriteHeader(http.StatusAccepted)
		_, err := ctx.Response.Write([]byte("1234567890"))
		if err != nil {
			println(err)
			ctx.Response.WriteHeader(http.StatusInternalServerError)
		}
	})

	go func() {
		err := server.Serve()
		a.ErrorType(err, http.ErrServerClosed, "assert.ErrorType 错误，%v", err)
		exit <- true
	}()
	time.Sleep(5000 * time.Microsecond) // 等待 go func() 完成

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost/root/m1/test").
		Do().
		Status(http.StatusAccepted)

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost/root/m2/test").
		Do().
		Status(http.StatusAccepted)

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost/root/mux/test").
		Do().
		Status(http.StatusAccepted)

	// static 中定义的静态文件
	server.Router().Static("/admin/{path}", "./testdata")
	rest.NewRequest(a, nil, http.MethodGet, "http://localhost/root/admin/file1.txt").
		Do().
		Status(http.StatusOK)

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost/root/admin/file1.txt").
		Do().
		Status(http.StatusOK)

	a.NotError(server.Close(0))
	<-exit
}

func TestServer_Close(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)
	exit := make(chan bool, 1)

	srv.Router().Get("/test", f202)
	srv.Router().Get("/close", func(ctx *Context) {
		_, err := ctx.Response.Write([]byte("closed"))
		if err != nil {
			ctx.Response.WriteHeader(http.StatusInternalServerError)
		}
		a.NotError(srv.Close(0))
	})

	go func() {
		err := srv.Serve()
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
		exit <- true
	}()

	// 等待 srv.Serve() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(5000 * time.Microsecond)

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost/root/test").
		Do().
		Status(http.StatusAccepted)

	// 连接被关闭，返回错误内容
	resp, err := http.Get("http://localhost/root/close")
	a.Error(err).Nil(resp)

	resp, err = http.Get("http://localhost/root/test")
	a.Error(err).Nil(resp)

	<-exit
}

func TestServer_CloseWithTimeout(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)
	exit := make(chan bool, 1)

	srv.Router().Get("/test", f202)
	srv.Router().Get("/close", func(ctx *Context) {
		ctx.Response.WriteHeader(http.StatusCreated)
		_, err := ctx.Response.Write([]byte("shutdown with ctx"))
		a.NotError(err)
		srv.Close(300 * time.Millisecond)
	})

	go func() {
		err := srv.Serve()
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
		exit <- true
	}()

	// 等待 srv.Serve() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(5000 * time.Microsecond)

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost/root/test").
		Do().
		Status(http.StatusAccepted)

	// 关闭指令可以正常执行
	rest.NewRequest(a, nil, http.MethodGet, "http://localhost/root/close").
		Do().
		Status(http.StatusCreated)

	// 未超时，但是拒绝新的链接
	resp, err := http.Get("http://localhost/root/test")
	a.Error(err).Nil(resp)

	// 已被关闭
	time.Sleep(30 * time.Microsecond)
	resp, err = http.Get("http://localhost/root/test")
	a.Error(err).Nil(resp)

	<-exit
}
