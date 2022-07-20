// SPDX-License-Identifier: MIT

package server_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"io/fs"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v2"
	"github.com/issue9/logs/v4"
	"github.com/issue9/mux/v7"
	"github.com/issue9/mux/v7/group"

	"github.com/issue9/web/serializer/text"
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

	router.Get("/m2/test", func(ctx *server.Context) server.Responser {
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

	cert, err := tls.LoadX509KeyPair("./testdata/cert.pem", "./testdata/key.pem")
	a.NotError(err).NotNil(cert)
	srv := servertest.NewTester(a, &server.Options{
		HTTPServer: &http.Server{
			Addr: ":8088",
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
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
	router.Get("/close", func(ctx *server.Context) server.Responser {
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
	router.Get("/close", func(ctx *server.Context) server.Responser {
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

func buildMiddleware(a *assert.Assertion, v string) server.Middleware {
	return server.MiddlewareFunc(func(next server.HandlerFunc) server.HandlerFunc {
		return func(ctx *server.Context) server.Responser {
			h := ctx.Header()
			val := h.Get("h")
			h.Set("h", v+val)

			resp := next(ctx)
			a.NotNil(resp)
			return resp
		}
	})
}

func TestMiddleware(t *testing.T) {
	a := assert.New(t, false)
	srv := servertest.NewTester(a, nil)
	count := 0

	router := srv.NewRouter(buildMiddleware(a, "b1"), buildMiddleware(a, "b2-"), server.MiddlewareFunc(func(next server.HandlerFunc) server.HandlerFunc {
		return func(ctx *server.Context) server.Responser {
			ctx.OnExit(func(status int) {
				count++
			})
			return next(ctx)
		}
	}))
	a.NotNil(router)
	router.Get("/path", servertest.BuildHandler(201))
	prefix := router.Prefix("/p1", buildMiddleware(a, "p1"), buildMiddleware(a, "p2-"))
	a.NotNil(prefix)
	prefix.Get("/path", servertest.BuildHandler(201))

	srv.GoServe()

	srv.Get("/p1/path").
		Header("accept", text.Mimetype).
		Do(nil).
		Status(http.StatusCreated).
		Header("h", "p1p2-b1b2-").
		StringBody("201")
	a.Equal(count, 1)

	srv.Get("/path").
		Header("accept", text.Mimetype).
		Do(nil).
		Status(http.StatusCreated).
		Header("h", "b1b2-").
		StringBody("201")
	a.Equal(count, 2)

	srv.Close(0)
	srv.Wait()
}

func TestServer_Routers(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewTester(a, nil)
	srv := s.Server()
	rs := srv.Routers()

	s.GoServe()

	ver := group.NewHeaderVersion("ver", "v", log.Default(), "2")
	a.NotNil(ver)
	r1 := rs.New("ver", ver, mux.URLDomain("https://example.com"))
	a.NotNil(r1)

	uu, err := r1.URL(false, "/posts/1", nil)
	a.NotError(err).Equal("https://example.com/posts/1", uu)

	r1.Prefix("/p1").Delete("/path", servertest.BuildHandler(http.StatusCreated))
	s.Delete("/p1/path").Header("Accept", "text/plain;v=2").Do(nil).Status(http.StatusCreated)

	r2 := rs.Router("ver")
	a.Equal(r2, r1)
	a.Equal(1, len(rs.Routers())).
		Equal(rs.Routers()[0].Name(), "ver")

	// 删除整个路由
	rs.Remove("ver")
	a.Equal(0, len(rs.Routers()))
	s.Delete("/p1/path").
		Header("Accept", "text/plain;v=2").
		Do(nil).
		Status(http.StatusNotFound)

	s.Close(0)
	s.Wait()
}

func TestServer_FileServer(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewTester(a, nil)
	rs := s.Server().Routers()

	s.GoServe()

	// 带版本

	ver := group.NewHeaderVersion("ver", "vv", log.Default(), "2")
	a.NotNil(ver)
	r := rs.New("ver", ver, mux.URLDomain("https://example.com/version"))
	r.Get("/ver/{path}", s.Server().FileServer(os.DirFS("./testdata"), "path", "index.html"))

	s.Get("/ver/file1.txt").
		Header("Accept", "text/plain;vv=2").
		Do(nil).
		Status(http.StatusOK).
		StringBody("file1")

	p := group.NewPathVersion("vv", "v2")
	a.NotNil(p)
	r = rs.New("path", p, mux.URLDomain("https://example.com/path"))
	r.Get("/path/{path}", s.Server().FileServer(os.DirFS("./testdata"), "path", "index.html"))

	s.Get("/v2/path/file1.txt").
		Do(nil).
		Status(http.StatusOK).
		StringBody("file1")

	r = s.NewRouter()
	r.Get("/m1/test", servertest.BuildHandler(201))
	r.Get("/client/{path}", s.Server().FileServer(os.DirFS("./testdata"), "path", "index.html"))

	s.Get("/m1/test").
		Header("accept", text.Mimetype).
		Do(nil).
		Status(http.StatusCreated).
		StringBody("201")

	// 定义的静态文件
	s.Get("/client/file1.txt").
		Do(nil).
		Status(http.StatusOK).
		Header("Content-Type", "text/plain; charset=utf-8").
		StringBody("file1")

	s.Get("/client/not-exists").
		Do(nil).
		Status(http.StatusNotFound)

	// 删除
	r.Remove("/client/{path}")
	s.Get("/client/file1.txt").
		Do(nil).
		Status(http.StatusNotFound)

	s.Close(0)
	s.Wait()
}

// 检测 204 是否存在 http: request method or response status code does not allow body
func TestContext_NoContent(t *testing.T) {
	a := assert.New(t, false)
	buf := new(bytes.Buffer)
	s := servertest.NewTester(a, &server.Options{HTTPServer: &http.Server{Addr: ":8080"}, Logs: logs.New(logs.NewTextWriter("15:04:05", buf))})

	s.NewRouter().Get("/204", func(ctx *server.Context) server.Responser {
		return ctx.NoContent()
	})

	s.GoServe()

	s.Get("/204").
		Header("Accept-Encoding", "gzip"). // 服务端不应该构建压缩对象
		Header("Accept", "application/json;charset=gbk").
		Do(nil).
		Status(http.StatusNoContent)

	s.Close(0)

	a.NotContains(buf.String(), "request method or response status code does not allow body")

	s.Wait()
}
