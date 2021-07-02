// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"io/fs"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
	"github.com/issue9/logs/v2"
	"github.com/issue9/mux/v5/group"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/web/content"
	"github.com/issue9/web/content/gob"
	"github.com/issue9/web/content/text"
)

var _ fs.FS = &Server{}

var f201 = func(ctx *Context) {
	ctx.Response.Header().Set("Content-Type", "text/html")
	ctx.Response.WriteHeader(http.StatusCreated)
	_, err := ctx.Response.Write([]byte("1234567890"))
	if err != nil {
		println(err)
	}
}

var f202 = func(ctx *Context) {
	ctx.Response.WriteHeader(http.StatusAccepted)
	_, err := ctx.Response.Write([]byte("1234567890"))
	if err != nil {
		println(err)
	}
}

func newLogs(a *assert.Assertion) *logs.Logs {
	l := logs.New()

	a.NotError(l.SetOutput(logs.LevelError, os.Stderr, "", 0))
	a.NotError(l.SetOutput(logs.LevelCritical, os.Stderr, "", 0))
	return l
}

// 声明一个 server 实例
func newServer(a *assert.Assertion) *Server {
	srv, err := New("app", "0.1.0", newLogs(a), &Options{Port: ":8080"})
	a.NotError(err).NotNil(srv)
	a.Equal(srv.Name(), "app").Equal(srv.Version(), "0.1.0")

	// srv.Catalog 默认指向 message.DefaultCatalog
	a.NotError(message.SetString(language.Und, "lang", "und"))
	a.NotError(message.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(message.SetString(language.TraditionalChinese, "lang", "hant"))

	mt := srv.Content()
	a.NotError(mt.AddMimetype("application/json", json.Marshal, json.Unmarshal))
	a.NotError(mt.AddMimetype("application/xml", xml.Marshal, xml.Unmarshal))
	a.NotError(mt.AddMimetype(content.DefaultMimetype, gob.Marshal, gob.Unmarshal))
	a.NotError(mt.AddMimetype(text.Mimetype, text.Marshal, text.Unmarshal))

	srv.Content().AddResult(411, 41110, "41110")

	return srv
}

func TestNewServer(t *testing.T) {
	a := assert.New(t)
	l := newLogs(a)

	srv, err := New("app", "0.1.0", l, nil)
	a.NotError(err).NotNil(srv)
	a.False(srv.Uptime().IsZero())
	a.Equal(l, srv.Logs())
	a.NotNil(srv.Cache())
	a.Equal(srv.Location(), time.Local)
	a.Equal(srv.httpServer.Handler, srv.groups)
	a.NotNil(srv.httpServer.BaseContext)
	a.Equal(srv.httpServer.Addr, "")
}

func TestGetServer(t *testing.T) {
	a := assert.New(t)
	type key int
	var k key = 0

	srv, err := New("app", "0.1.0", newLogs(a), nil)
	a.NotError(err).NotNil(srv)
	err = srv.content.AddMimetype(text.Mimetype, text.Marshal, text.Unmarshal)
	a.NotError(err)
	var isRequested bool

	router, err := srv.NewRouter("default", "http://localhost:8081/", group.MatcherFunc(group.Any))
	a.NotError(err).NotNil(router)
	router.MuxRouter().GetFunc("/path", func(w http.ResponseWriter, r *http.Request) {
		s1 := GetServer(r)
		a.NotNil(s1).Equal(s1, srv)

		v := r.Context().Value(k)
		a.Nil(v)

		ctx1 := NewContext(w, r)
		a.NotNil(ctx1)
		ctx2 := NewContext(w, ctx1.Request)
		a.Equal(ctx1, ctx2)

		isRequested = true
	})

	go func() {
		a.Equal(srv.Serve(), http.ErrServerClosed)
	}()
	time.Sleep(500 * time.Millisecond)
	rest.NewRequest(a, nil, http.MethodGet, "http://localhost/path").
		Header("Accept", text.Mimetype).
		Do().
		Success("未正确返回状态码")

	// 不是从 Server 生成的 *http.Request，则会 panic
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	a.Panic(func() {
		GetServer(r)
	})

	a.NotError(srv.Close(0))
	a.True(isRequested, "未正常访问 /path")

	// BaseContext

	srv, err = New("app", "0.1.0", newLogs(a), &Options{
		HTTPServer: func(s *http.Server) {
			s.BaseContext = func(n net.Listener) context.Context {
				return context.WithValue(context.Background(), k, 1)
			}
		},
	})
	a.NotError(err).NotNil(srv)

	isRequested = false
	router, err = srv.NewRouter("default", "http://localhost:8080/", group.MatcherFunc(group.Any))
	a.NotError(err).NotNil(router)
	router.MuxRouter().GetFunc("/path", func(w http.ResponseWriter, r *http.Request) {
		s1 := GetServer(r)
		a.NotNil(s1).Equal(s1, srv)

		v := r.Context().Value(k) // BaseContext 中设置了 k 的值
		a.Equal(v, 1)

		isRequested = true
	})
	go func() {
		a.Equal(srv.Serve(), http.ErrServerClosed)
	}()
	time.Sleep(500 * time.Millisecond)
	rest.NewRequest(a, nil, http.MethodGet, "http://localhost/path").Do().Success()
	a.NotError(srv.Close(0))
	a.True(isRequested, "未正常访问 /path")
}

func TestServer_vars(t *testing.T) {
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

	v11, found := srv.Get(v1)
	a.True(found).Equal(v11, 1)
	v22, found := srv.Get(v2)
	a.True(found).Equal(v22, 3)
}

func TestServer_Serve(t *testing.T) {
	a := assert.New(t)
	exit := make(chan bool, 1)

	server := newServer(a)
	router, err := server.NewRouter("default", "http://localhost:8080/root/", group.MatcherFunc(group.Any))
	a.NotError(err).NotNil(router)
	router.Get("/mux/test", f202)

	m1 := server.NewModule("m1", "m1 desc")
	a.NotNil(m1)
	m1.AddInit("init", func() error {
		router.Get("/m1/test", f202)
		return nil
	})
	m1.NewTag("tag1")

	m2 := server.NewModule("m2", "m2 desc", "m1")
	a.NotNil(m2)
	m2.AddInit("init m2", func() error {
		router.Get("/m2/test", func(ctx *Context) {
			srv := ctx.Server()
			a.NotNil(srv)
			a.Equal(2, len(srv.Modules()))
			a.Equal(srv.Tags(), []string{"tag1"})

			ctx.Response.WriteHeader(http.StatusAccepted)
			_, err := ctx.Response.Write([]byte("1234567890"))
			if err != nil {
				println(err)
			}

			// 动态加载模块
			m3 := server.NewModule("m3", "m3 desc", "m1")
			a.NotNil(m3)
			m3.AddInit("init3", func() error { return nil })
			a.NotError(server.AddModule(m3))

			a.Equal(3, len(srv.Modules()))
			a.True(m3.Inited())
		})
		return nil
	})

	a.NotError(server.AddModule(m1, m2))

	go func() {
		err := server.Serve()
		a.ErrorIs(err, http.ErrServerClosed, "assert.ErrorType 错误，%v", err)
		exit <- true
	}()
	time.Sleep(5000 * time.Microsecond) // 等待 go func() 完成

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8080/root/m1/test").
		Do().
		Status(http.StatusAccepted)

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8080/root/m2/test").
		Do().
		Status(http.StatusAccepted)

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8080/root/mux/test").
		Do().
		Status(http.StatusAccepted)

	// static 中定义的静态文件
	a.NotError(router.Static("/admin/{path}", "./testdata", "index.html"))
	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8080/root/admin/file1.txt").
		Do().
		Status(http.StatusOK)

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8080/root/admin/file1.txt").
		Do().
		Status(http.StatusOK)

	a.NotError(server.Close(0))
	<-exit
}

func TestServer_Serve_HTTPS(t *testing.T) {
	a := assert.New(t)
	exit := make(chan bool, 1)

	server, err := New("app", "0.1.0", newLogs(a), &Options{
		Port: ":8088",
		HTTPServer: func(srv *http.Server) {
			cert, err := tls.LoadX509KeyPair("./testdata/cert.pem", "./testdata/key.pem")
			a.NotError(err).NotNil(cert)
			srv.TLSConfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
		},
	})
	a.NotError(err).NotNil(server)
	err = server.content.AddMimetype(text.Mimetype, text.Marshal, text.Unmarshal)
	a.NotError(err)

	router, err := server.NewRouter("default", "https://localhost/api", group.MatcherFunc(group.Any))
	a.NotError(err).NotNil(router)
	router.Get("/mux/test", f202)

	go func() {
		err := server.Serve()
		a.ErrorType(err, http.ErrServerClosed, "assert.ErrorType 错误，%v", err)
		exit <- true
	}()
	time.Sleep(5000 * time.Microsecond) // 等待 go func() 完成

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	rest.NewRequest(a, client, http.MethodGet, "https://localhost:8088/api/mux/test").
		Do().
		Status(http.StatusAccepted)

	// 无效的 http 请求
	rest.NewRequest(a, client, http.MethodGet, "http://localhost:8088/api/mux/test").
		Do().
		Status(http.StatusBadRequest)
	rest.NewRequest(a, client, http.MethodGet, "http://localhost:8088/api/mux").
		Do().
		Status(http.StatusBadRequest)

	a.NotError(server.Close(0))
	<-exit
}

func TestServer_Close(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)
	exit := make(chan bool, 1)
	router, err := srv.NewRouter("default", "https://localhost:8088/root", group.MatcherFunc(group.Any))
	a.NotError(err).NotNil(router)

	router.Get("/test", f202)
	router.Get("/close", func(ctx *Context) {
		_, err := ctx.Response.Write([]byte("closed"))
		if err != nil {
			ctx.Response.WriteHeader(http.StatusInternalServerError)
		}
		a.NotError(srv.Close(0))
	})

	go func() {
		a.ErrorIs(srv.Serve(), http.ErrServerClosed)
		exit <- true
	}()

	// 等待 srv.Serve() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(500 * time.Millisecond)

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8080/root/test").
		Do().
		Status(http.StatusAccepted)

	// 连接被关闭，返回错误内容
	resp, err := http.Get("http://localhost:8080/root/close")
	a.Error(err).Nil(resp)

	resp, err = http.Get("http://localhost:8080/root/test")
	a.Error(err).Nil(resp)

	<-exit
}

func TestServer_CloseWithTimeout(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)
	exit := make(chan bool, 1)
	router, err := srv.NewRouter("default", "https://localhost:8088/root", group.MatcherFunc(group.Any))
	a.NotError(err).NotNil(router)

	router.Get("/test", f202)
	router.Get("/close", func(ctx *Context) {
		ctx.Response.WriteHeader(http.StatusCreated)
		_, err := ctx.Response.Write([]byte("shutdown with ctx"))
		a.NotError(err)
		a.NotError(srv.Close(300 * time.Millisecond))
	})

	go func() {
		err := srv.Serve()
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
		exit <- true
	}()

	// 等待 srv.Serve() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(5000 * time.Microsecond)

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8080/root/test").
		Do().
		Status(http.StatusAccepted)

	// 关闭指令可以正常执行
	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8080/root/close").
		Do().
		Status(http.StatusCreated)

	// 未超时，但是拒绝新的链接
	resp, err := http.Get("http://localhost:8080/root/test")
	a.Error(err).Nil(resp)

	// 已被关闭
	time.Sleep(30 * time.Microsecond)
	resp, err = http.Get("http://localhost:8080/root/test")
	a.Error(err).Nil(resp)

	<-exit
}

func TestServer_DisableCompression(t *testing.T) {
	a := assert.New(t)
	server := newServer(a)
	srv := rest.NewServer(t, server.groups, nil)
	defer srv.Close()
	router, err := server.NewRouter("default", "http://localhost:8081/root", group.MatcherFunc(group.Any))
	a.NotError(err).NotNil(router)

	a.NotError(router.Static("/client/{path}", "./testdata/", "index.html"))

	srv.Get("/root/client/file1.txt").
		Header("Accept-Encoding", "gzip,deflate;q=0.8").
		Do().
		Status(http.StatusOK).
		Header("Content-Type", "text/plain; charset=utf-8").
		Header("Content-Encoding", "gzip").
		Header("Vary", "Content-Encoding")

	srv.Get("/root/client/file1.txt").
		Do().
		Status(http.StatusOK).
		Header("Content-Type", "text/plain; charset=utf-8").
		Header("Content-Encoding", "").
		Header("Vary", "Content-Encoding")

	server.DisableCompression(true)
	srv.Get("/root/client/file1.txt").
		Header("Accept-Encoding", "gzip,deflate;q=0.8").
		Do().
		Status(http.StatusOK).
		Header("Content-Type", "text/plain; charset=utf-8").
		Header("Content-Encoding", "").
		Header("Vary", "")
}
