// SPDX-License-Identifier: MIT

package server_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"io/fs"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v2"

	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

var _ fs.FS = &server.Server{}

func TestServer_Vars(t *testing.T) {
	a := assert.New(t, false)
	srv, err := server.New("app", "1.0.0", nil)
	a.NotError(err).NotNil(srv)

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

	srv := servertest.NewTester(a, nil)
	router := srv.NewRouter()
	router.Get("/mux/test", servertest.BuildHandler(202))
	router.Get("/m1/test", servertest.BuildHandler(202))

	a.False(srv.Server().Serving())

	router.Get("/m2/test", func(ctx *server.Context) *server.Response {
		a.True(srv.Server().Serving())

		srv := ctx.Server()
		a.NotNil(srv)

		ctx.WriteHeader(http.StatusAccepted)
		_, err := ctx.Write([]byte("1234567890"))
		if err != nil {
			println(err)
		}
		return nil
	})

	srv.GoServe()

	srv.Get("http://localhost:8080/m1/test").Do(nil).Status(http.StatusAccepted)

	srv.Get("http://localhost:8080/m2/test").Do(nil).Status(http.StatusAccepted)

	srv.Get("http://localhost:8080/mux/test").Do(nil).Status(http.StatusAccepted)

	// static 中定义的静态文件
	router.Get("/admin/{path}", srv.Server().FileServer(os.DirFS("./testdata"), "path", "index.html"))
	srv.Get("http://localhost:8080/admin/file1.txt").Do(nil).Status(http.StatusOK)

	srv.Close(0)
	srv.Wait()

	a.False(srv.Server().Serving())
}

func TestServer_Serve_HTTPS(t *testing.T) {
	a := assert.New(t, false)

	srv := servertest.NewTester(a, &server.Options{
		Port: ":8088",
		HTTPServer: func(srv *http.Server) {
			cert, err := tls.LoadX509KeyPair("./testdata/cert.pem", "./testdata/key.pem")
			a.NotError(err).NotNil(cert)
			srv.TLSConfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
		},
	})

	router := srv.NewRouter()
	router.Get("/mux/test", servertest.BuildHandler(202))

	srv.GoServe()

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	resp, err := client.Get("https://localhost:8088/mux/test")
	a.NotError(err).Equal(resp.StatusCode, http.StatusAccepted)

	// 无效的 http 请求
	resp, err = client.Get("http://localhost:8088/mux/test")
	a.NotError(err).Equal(resp.StatusCode, http.StatusBadRequest)

	resp, err = client.Get("http://localhost:8088/mux")
	a.NotError(err).Equal(resp.StatusCode, http.StatusBadRequest)

	srv.Close(0)
	srv.Wait()
}

func TestServer_Close(t *testing.T) {
	a := assert.New(t, false)
	srv := servertest.NewTester(a, nil)
	router := srv.NewRouter()

	router.Get("/test", servertest.BuildHandler(202))
	router.Get("/close", func(ctx *server.Context) *server.Response {
		_, err := ctx.Write([]byte("closed"))
		if err != nil {
			ctx.WriteHeader(http.StatusInternalServerError)
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

	srv.Get("http://localhost:8080/test").Do(nil).Status(http.StatusAccepted)

	// 连接被关闭，返回错误内容
	resp, err := http.Get("http://localhost:8080/close")
	a.Error(err).Nil(resp)

	resp, err = http.Get("http://localhost:8080/test")
	a.Error(err).Nil(resp)

	str := buf.String()
	a.Contains(str, "canceled").
		Contains(str, "RegisterOnClose")

	srv.Close(0)
	srv.Wait()
}

func TestServer_CloseWithTimeout(t *testing.T) {
	a := assert.New(t, false)
	srv := servertest.NewTester(a, nil)
	router := srv.NewRouter()

	router.Get("/test", servertest.BuildHandler(202))
	router.Get("/close", func(ctx *server.Context) *server.Response {
		ctx.WriteHeader(http.StatusCreated)
		_, err := ctx.Write([]byte("shutdown with ctx"))
		a.NotError(err)
		a.NotError(srv.Server().Close(300 * time.Millisecond))

		return nil
	})

	srv.GoServe()

	srv.Get("http://localhost:8080/test").Do(nil).Status(http.StatusAccepted)

	// 关闭指令可以正常执行
	srv.Get("http://localhost:8080/close").Do(nil).Status(http.StatusCreated)

	// 未超时，但是拒绝新的链接
	resp, err := http.Get("http://localhost:8080/test")
	a.Error(err).Nil(resp)

	// 已被关闭
	time.Sleep(30 * time.Microsecond)
	resp, err = http.Get("http://localhost:8080/test")
	a.Error(err).Nil(resp)

	srv.Close(0)
	srv.Wait()
}
