// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"io/fs"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v3"
	"github.com/issue9/mux/v5/group"
	"golang.org/x/text/language"

	"github.com/issue9/web/serialization/gob"
	"github.com/issue9/web/serialization/text"
)

var _ fs.FS = &Server{}

var f201 = func(ctx *Context) Responser {
	ctx.Response.Header().Set("Content-Type", "text/html")
	ctx.Response.WriteHeader(http.StatusCreated)
	_, err := ctx.Response.Write([]byte("1234567890"))
	if err != nil {
		println(err)
	}

	return nil
}

var f202 = func(ctx *Context) Responser {
	ctx.Response.WriteHeader(http.StatusAccepted)
	_, err := ctx.Response.Write([]byte("1234567890"))
	if err != nil {
		println(err)
	}

	return nil
}

// 声明一个 server 实例
func newServer(a *assert.Assertion, o *Options) *Server {
	if o == nil {
		o = &Options{Port: ":8080"}
	}
	if o.Logs == nil { // 默认重定向到 os.Stderr
		l, err := logs.New(nil)
		a.NotError(err).NotNil(l)

		a.NotError(l.SetOutput(logs.LevelDebug, os.Stderr))
		a.NotError(l.SetOutput(logs.LevelError, os.Stderr))
		a.NotError(l.SetOutput(logs.LevelCritical, os.Stderr))
		a.NotError(l.SetOutput(logs.LevelInfo, os.Stdout))
		a.NotError(l.SetOutput(logs.LevelTrace, os.Stdout))
		a.NotError(l.SetOutput(logs.LevelWarn, os.Stdout))
		o.Logs = l
	}

	srv, err := New("app", "0.1.0", o)
	a.NotError(err).NotNil(srv)
	a.Equal(srv.Name(), "app").Equal(srv.Version(), "0.1.0")

	// locale
	b := srv.Locale().Builder()
	a.NotError(b.SetString(language.Und, "lang", "und"))
	a.NotError(b.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(b.SetString(language.TraditionalChinese, "lang", "hant"))

	// mimetype
	a.NotError(srv.Mimetypes().Add(json.Marshal, json.Unmarshal, "application/json"))
	a.NotError(srv.Mimetypes().Add(xml.Marshal, xml.Unmarshal, "application/xml"))
	a.NotError(srv.Mimetypes().Add(gob.Marshal, gob.Unmarshal, DefaultMimetype))
	a.NotError(srv.Mimetypes().Add(text.Marshal, text.Unmarshal, text.Mimetype))

	srv.AddResult(411, "41110", localeutil.Phrase("41110"))

	return srv
}

func TestNewServer(t *testing.T) {
	a := assert.New(t, false)

	srv, err := New("app", "0.1.0", nil)
	a.NotError(err).NotNil(srv)
	a.False(srv.Uptime().IsZero())
	a.NotNil(srv.Cache())
	a.Equal(srv.Location(), time.Local)
	a.Equal(srv.httpServer.Handler, srv.group)
	a.NotNil(srv.httpServer.BaseContext)
	a.Equal(srv.httpServer.Addr, "")
}

func TestGetServer(t *testing.T) {
	a := assert.New(t, false)
	type key int
	var k key = 0

	srv := newServer(a, &Options{Port: ":8080"})
	var isRequested bool

	router := srv.NewRouter("default", "http://localhost:8081/", group.MatcherFunc(group.Any))
	a.NotNil(router)
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
	r, err := http.NewRequest(http.MethodGet, "http://localhost:8080/path", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", text.Mimetype)
	resp, err := http.DefaultClient.Do(r)
	a.NotError(err).Equal(resp.StatusCode, 200)

	// 不是从 Server 生成的 *http.Request，则会 panic
	r, err = http.NewRequest(http.MethodGet, "/path", nil)
	a.NotError(err).NotNil(r)
	a.Panic(func() {
		GetServer(r)
	})

	a.NotError(srv.Close(0))
	a.True(isRequested, "未正常访问 /path")

	// BaseContext

	srv = newServer(a, &Options{
		Port: ":8080",
		HTTPServer: func(s *http.Server) {
			s.BaseContext = func(n net.Listener) context.Context {
				return context.WithValue(context.Background(), k, 1)
			}
		},
	})

	isRequested = false
	router = srv.NewRouter("default", "http://localhost:8080/", group.MatcherFunc(group.Any))
	a.NotNil(router)
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
	resp, err = http.Get("http://localhost:8080/path")
	a.NotError(err).Equal(resp.StatusCode, http.StatusOK)

	a.NotError(srv.Close(0))
	a.True(isRequested, "未正常访问 /path")
}

func TestServer_Vars(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)

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

	srv.Vars().Store(v1, 1)
	srv.Vars().Store(v2, 2)
	srv.Vars().Store(v3, 3)

	v11, found := srv.Vars().Load(v1)
	a.True(found).Equal(v11, 1)
	v22, found := srv.Vars().Load(v2)
	a.True(found).Equal(v22, 3)
}

func TestServer_Serve(t *testing.T) {
	a := assert.New(t, false)
	exit := make(chan bool, 1)

	srv := newServer(a, nil)
	router := srv.NewRouter("default", "http://localhost:8080/", group.MatcherFunc(group.Any))
	a.NotNil(router)
	router.Get("/mux/test", f202)
	router.Get("/m1/test", f202)

	a.False(srv.Serving())

	router.Get("/m2/test", func(ctx *Context) Responser {
		a.True(srv.Serving())

		srv := ctx.Server()
		a.NotNil(srv)

		ctx.Response.WriteHeader(http.StatusAccepted)
		_, err := ctx.Response.Write([]byte("1234567890"))
		if err != nil {
			println(err)
		}
		return nil
	})

	go func() {
		err := srv.Serve()
		a.ErrorIs(err, http.ErrServerClosed, "assert.ErrorType 错误，%v", err)
		exit <- true
	}()
	time.Sleep(5000 * time.Microsecond) // 等待 go func() 完成

	resp, err := http.Get("http://localhost:8080/m1/test")
	a.NotError(err).Equal(resp.StatusCode, http.StatusAccepted)

	resp, err = http.Get("http://localhost:8080/m2/test")
	a.NotError(err).Equal(resp.StatusCode, http.StatusAccepted)

	resp, err = http.Get("http://localhost:8080/mux/test")
	a.NotError(err).Equal(resp.StatusCode, http.StatusAccepted)

	// static 中定义的静态文件
	router.Static("/admin/{path}", "./testdata", "index.html")
	resp, err = http.Get("http://localhost:8080/admin/file1.txt")
	a.NotError(err).Equal(resp.StatusCode, http.StatusOK)

	a.NotError(srv.Close(0))
	<-exit

	a.False(srv.Serving())
}

func TestServer_Serve_HTTPS(t *testing.T) {
	a := assert.New(t, false)
	exit := make(chan bool, 1)

	srv := newServer(a, &Options{
		Port: ":8088",
		HTTPServer: func(srv *http.Server) {
			cert, err := tls.LoadX509KeyPair("./testdata/cert.pem", "./testdata/key.pem")
			a.NotError(err).NotNil(cert)
			srv.TLSConfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
		},
	})

	router := srv.NewRouter("default", "https://localhost/root", group.MatcherFunc(group.Any))
	a.NotNil(router)
	router.Get("/mux/test", f202)

	go func() {
		err := srv.Serve()
		a.ErrorIs(err, http.ErrServerClosed, "assert.ErrorType 错误，%v", err)
		exit <- true
	}()
	time.Sleep(5000 * time.Microsecond) // 等待 go func() 完成

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get("https://localhost:8088/mux/test")
	a.NotError(err).Equal(resp.StatusCode, http.StatusAccepted)

	// 无效的 http 请求
	resp, err = client.Get("http://localhost:8088/mux/test")
	a.NotError(err).Equal(resp.StatusCode, http.StatusBadRequest)

	resp, err = client.Get("http://localhost:8088/mux")
	a.NotError(err).Equal(resp.StatusCode, http.StatusBadRequest)

	a.NotError(srv.Close(0))
	<-exit
}

func TestServer_Close(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)
	exit := make(chan bool, 1)
	router := srv.NewRouter("default", "https://localhost:8088/root", group.MatcherFunc(group.Any))
	a.NotNil(router)

	router.Get("/test", f202)
	router.Get("/close", func(ctx *Context) Responser {
		_, err := ctx.Response.Write([]byte("closed"))
		if err != nil {
			ctx.Response.WriteHeader(http.StatusInternalServerError)
		}
		a.NotError(srv.Close(0))

		return nil
	})

	buf := new(bytes.Buffer)
	srv.AddService("srv1", func(ctx context.Context) error {
		c := time.Tick(10 * time.Millisecond)
		for {
			select {
			case <-c:
				println("TestServer_Close tick...")
			case <-ctx.Done():
				buf.WriteString("canceled\n")
				println("TestServer_Close canceled...")
				return context.Canceled
			}
		}
	})

	srv.OnClose(func() error {
		buf.WriteString("RegisterOnClose\n")
		println("TestServer_Close RegisterOnClose...")
		return nil
	})

	go func() {
		a.ErrorIs(srv.Serve(), http.ErrServerClosed)
		exit <- true
	}()

	// 等待 srv.Serve() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(500 * time.Millisecond)

	resp, err := http.Get("http://localhost:8080/test")
	a.NotError(err).Equal(resp.StatusCode, http.StatusAccepted)

	// 连接被关闭，返回错误内容
	resp, err = http.Get("http://localhost:8080/close")
	a.Error(err).Nil(resp)

	resp, err = http.Get("http://localhost:8080/test")
	a.Error(err).Nil(resp)

	str := buf.String()
	a.Contains(str, "canceled").
		Contains(str, "RegisterOnClose")

	<-exit
}

func TestServer_CloseWithTimeout(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)
	exit := make(chan bool, 1)
	router := srv.NewRouter("default", "https://localhost:8088/root", group.MatcherFunc(group.Any))
	a.NotNil(router)

	router.Get("/test", f202)
	router.Get("/close", func(ctx *Context) Responser {
		ctx.Response.WriteHeader(http.StatusCreated)
		_, err := ctx.Response.Write([]byte("shutdown with ctx"))
		a.NotError(err)
		a.NotError(srv.Close(300 * time.Millisecond))

		return nil
	})

	go func() {
		err := srv.Serve()
		a.Error(err).ErrorIs(err, http.ErrServerClosed, "错误信息为:%v", err)
		exit <- true
	}()

	// 等待 srv.Serve() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(5000 * time.Microsecond)

	resp, err := http.Get("http://localhost:8080/test")
	a.NotError(err).Equal(resp.StatusCode, http.StatusAccepted)

	// 关闭指令可以正常执行
	resp, err = http.Get("http://localhost:8080/close")
	a.NotError(err).Equal(resp.StatusCode, http.StatusCreated)

	// 未超时，但是拒绝新的链接
	resp, err = http.Get("http://localhost:8080/test")
	a.Error(err).Nil(resp)

	// 已被关闭
	time.Sleep(30 * time.Microsecond)
	resp, err = http.Get("http://localhost:8080/test")
	a.Error(err).Nil(resp)

	<-exit
}

func TestServer_DisableCompression(t *testing.T) {
	a := assert.New(t, false)
	server := newServer(a, nil)
	srv := rest.NewServer(a, server.group, nil)

	router := server.NewRouter("default", "http://localhost:8081/root", group.MatcherFunc(group.Any))
	a.NotNil(router)

	router.Static("/client/{path}", "./testdata/", "index.html")

	srv.Get("/client/file1.txt").
		Header("Accept-Encoding", "gzip,deflate;q=0.8").
		Do(nil).
		Status(http.StatusOK).
		Header("Content-Type", "text/plain; charset=utf-8").
		Header("Content-Encoding", "gzip").
		Header("Vary", "Content-Encoding")

	srv.Get("/client/file1.txt").
		Do(nil).
		Status(http.StatusOK).
		Header("Content-Type", "text/plain; charset=utf-8").
		Header("Content-Encoding", "").
		Header("Vary", "Content-Encoding")

	server.DisableCompression(true)
	srv.Get("/client/file1.txt").
		Header("Accept-Encoding", "gzip,deflate;q=0.8").
		Do(nil).
		Status(http.StatusOK).
		Header("Content-Type", "text/plain; charset=utf-8").
		Header("Content-Encoding", "").
		Header("Vary", "")
}
