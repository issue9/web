// SPDX-License-Identifier: MIT

package server_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"io/fs"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v2"
	"github.com/issue9/mux/v6/group"

	"github.com/issue9/web/serialization/text"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

var (
	_ fs.FS = &server.Server{}

	// 需要 accept 为 text/plian 否则可能输出内容会有误。
	f202 = func(ctx *server.Context) server.Responser {
		return server.Object(http.StatusAccepted, []byte("1234567890"), nil)
	}
)

func TestGetServer(t *testing.T) {
	a := assert.New(t, false)
	type key int
	var k key = 0

	srv := servertest.NewServer(a, &server.Options{Port: ":8080"})
	var isRequested bool

	router := srv.Server().NewRouter("default", "http://localhost:8081/", group.MatcherFunc(group.Any))
	a.NotNil(router)
	router.Get("/path", func(ctx *server.Context) server.Responser {
		s1 := server.GetServer(ctx.Request)
		a.NotNil(s1).Equal(s1, srv.Server())

		v := ctx.Request.Context().Value(k)
		a.Nil(v)

		isRequested = true

		return nil
	})

	srv.GoServe()

	r, err := http.NewRequest(http.MethodGet, "http://localhost:8080/path", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", text.Mimetype)
	resp, err := http.DefaultClient.Do(r)
	a.NotError(err).Equal(resp.StatusCode, 200)

	// 不是从 Server 生成的 *http.Request，则会 panic
	r, err = http.NewRequest(http.MethodGet, "/path", nil)
	a.NotError(err).NotNil(r)
	a.Panic(func() {
		server.GetServer(r)
	})

	a.NotError(srv.Server().Close(0))
	a.True(isRequested, "未正常访问 /path")

	srv.Wait()

	// BaseContext

	srv = servertest.NewServer(a, &server.Options{
		Port: ":8080",
		HTTPServer: func(s *http.Server) {
			s.BaseContext = func(n net.Listener) context.Context {
				return context.WithValue(context.Background(), k, 1)
			}
		},
	})

	isRequested = false
	router = srv.Server().NewRouter("default", "http://localhost:8080/", group.MatcherFunc(group.Any))
	a.NotNil(router)
	router.Get("/path", func(ctx *server.Context) server.Responser {
		s1 := server.GetServer(ctx.Request)
		a.NotNil(s1).Equal(s1, srv.Server())

		v := ctx.Request.Context().Value(k) // BaseContext 中设置了 k 的值
		a.Equal(v, 1)

		isRequested = true

		return nil
	})
	srv.GoServe()
	resp, err = http.Get("http://localhost:8080/path")
	a.NotError(err).Equal(resp.StatusCode, http.StatusOK)

	a.NotError(srv.Server().Close(0))
	a.True(isRequested, "未正常访问 /path")

	srv.Wait()
}

func TestServer_Vars(t *testing.T) {
	a := assert.New(t, false)
	srv := servertest.NewServer(a, nil)

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

	srv.Server().Vars().Store(v1, 1)
	srv.Server().Vars().Store(v2, 2)
	srv.Server().Vars().Store(v3, 3)

	v11, found := srv.Server().Vars().Load(v1)
	a.True(found).Equal(v11, 1)
	v22, found := srv.Server().Vars().Load(v2)
	a.True(found).Equal(v22, 3)
}

func TestServer_Serve(t *testing.T) {
	a := assert.New(t, false)

	srv := servertest.NewServer(a, nil)
	router := srv.Server().NewRouter("default", "http://localhost:8080/", group.MatcherFunc(group.Any))
	a.NotNil(router)
	router.Get("/mux/test", f202)
	router.Get("/m1/test", f202)

	a.False(srv.Server().Serving())

	router.Get("/m2/test", func(ctx *server.Context) server.Responser {
		a.True(srv.Server().Serving())

		srv := ctx.Server()
		a.NotNil(srv)

		ctx.Response.WriteHeader(http.StatusAccepted)
		_, err := ctx.Response.Write([]byte("1234567890"))
		if err != nil {
			println(err)
		}
		return nil
	})

	srv.GoServe()

	resp, err := http.Get("http://localhost:8080/m1/test")
	a.NotError(err).Equal(resp.StatusCode, http.StatusAccepted)

	resp, err = http.Get("http://localhost:8080/m2/test")
	a.NotError(err).Equal(resp.StatusCode, http.StatusAccepted)

	resp, err = http.Get("http://localhost:8080/mux/test")
	a.NotError(err).Equal(resp.StatusCode, http.StatusAccepted)

	// static 中定义的静态文件
	router.Get("/admin/{path}", srv.Server().FileServer(os.DirFS("./testdata"), "path", "index.html"))
	resp, err = http.Get("http://localhost:8080/admin/file1.txt")
	a.NotError(err).Equal(resp.StatusCode, http.StatusOK)

	a.NotError(srv.Server().Close(0))
	srv.Wait()

	a.False(srv.Server().Serving())
}

func TestServer_Serve_HTTPS(t *testing.T) {
	a := assert.New(t, false)

	srv := servertest.NewServer(a, &server.Options{
		Port: ":8088",
		HTTPServer: func(srv *http.Server) {
			cert, err := tls.LoadX509KeyPair("./testdata/cert.pem", "./testdata/key.pem")
			a.NotError(err).NotNil(cert)
			srv.TLSConfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
		},
	})

	router := srv.Server().NewRouter("default", "https://localhost/root", group.MatcherFunc(group.Any))
	a.NotNil(router)
	router.Get("/mux/test", f202)

	srv.GoServe()

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

	a.NotError(srv.Server().Close(0))
	srv.Wait()
}

func TestServer_Close(t *testing.T) {
	a := assert.New(t, false)
	srv := servertest.NewServer(a, nil)
	router := srv.Server().NewRouter("default", "https://localhost:8088/root", group.MatcherFunc(group.Any))
	a.NotNil(router)

	router.Get("/test", f202)
	router.Get("/close", func(ctx *server.Context) server.Responser {
		_, err := ctx.Response.Write([]byte("closed"))
		if err != nil {
			ctx.Response.WriteHeader(http.StatusInternalServerError)
		}
		a.NotError(srv.Server().Close(0))

		return nil
	})

	buf := new(bytes.Buffer)
	srv.Server().AddService("srv1", func(ctx context.Context) error {
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

	srv.Server().OnClose(func() error {
		buf.WriteString("RegisterOnClose\n")
		println("TestServer_Close RegisterOnClose...")
		return nil
	})

	srv.GoServe()

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

	srv.Wait()
}

func TestServer_CloseWithTimeout(t *testing.T) {
	a := assert.New(t, false)
	srv := servertest.NewServer(a, nil)
	router := srv.Server().NewRouter("default", "https://localhost:8088/root", group.MatcherFunc(group.Any))
	a.NotNil(router)

	router.Get("/test", f202)
	router.Get("/close", func(ctx *server.Context) server.Responser {
		ctx.Response.WriteHeader(http.StatusCreated)
		_, err := ctx.Response.Write([]byte("shutdown with ctx"))
		a.NotError(err)
		a.NotError(srv.Server().Close(300 * time.Millisecond))

		return nil
	})

	srv.GoServe()

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

	srv.Wait()
}
