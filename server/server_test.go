// SPDX-License-Identifier: MIT

package server

import (
	"crypto/tls"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"golang.org/x/text/language"

	"github.com/issue9/web"
	"github.com/issue9/web/cache"
	"github.com/issue9/web/codec"
	"github.com/issue9/web/codec/mimetype/json"
	"github.com/issue9/web/codec/mimetype/xml"
	"github.com/issue9/web/logs"
	"github.com/issue9/web/servertest"
)

var (
	_ servertest.Server = &httpServer{}
	_ web.Server        = &httpServer{}
)

func buildHandler(code int) web.HandlerFunc {
	return func(ctx *web.Context) web.Responser {
		return web.ResponserFunc(func(ctx *web.Context) web.Problem {
			ctx.Render(code, code)
			return nil
		})
	}
}

func TestNewServer(t *testing.T) {
	a := assert.New(t, false)

	srv, err := New("app", "0.1.0", nil)
	s := srv.(*httpServer)
	a.NotError(err).NotNil(srv).
		False(srv.Uptime().IsZero()).
		NotNil(srv.Cache()).
		Equal(srv.Location(), time.Local).
		Equal(s.httpServer.Handler, s.routers).
		Equal(s.httpServer.Addr, "")

	d, ok := srv.Cache().(cache.Driver)
	a.True(ok).
		NotNil(d).
		NotNil(d.Driver())
	a.True(srv.Codec().CanCompress())
	srv.Codec().SetCompress(false)
	a.False(srv.Codec().CanCompress())
}

func newTestServer(a *assert.Assertion, o *Options) *httpServer {
	if o == nil {
		o = &Options{HTTPServer: &http.Server{Addr: ":8080"}, Language: language.English} // 指定不存在的语言
	}
	if o.Logs == nil { // 默认重定向到 os.Stderr
		o.Logs = &logs.Options{
			Handler:  logs.NewTermHandler(os.Stderr, nil),
			Location: true,
			Created:  logs.NanoLayout,
			Levels:   logs.AllLevels(),
		}
	}
	if o.Compressions == nil {
		o.Compressions = codec.DefaultCompressions()
	}
	if o.Mimetypes == nil {
		o.Mimetypes = []*codec.Mimetype{
			{Name: "application/json", MarshalBuilder: json.BuildMarshal, Unmarshal: json.Unmarshal, Problem: "application/problem+json"},
			{Name: "application/xml", MarshalBuilder: xml.BuildMarshal, Unmarshal: xml.Unmarshal, Problem: ""},
			{Name: "nil", MarshalBuilder: nil, Unmarshal: nil, Problem: ""},
		}
	}

	srv, err := New("app", "0.1.0", o)
	a.NotError(err).NotNil(srv)
	a.Equal(srv.Name(), "app").Equal(srv.Version(), "0.1.0")

	// locale
	b := srv.Catalog()
	a.NotError(b.SetString(language.Und, "lang", "und"))
	a.NotError(b.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(b.SetString(language.TraditionalChinese, "lang", "hant"))

	srv.AddProblem("41110", 411, web.Phrase("lang"), web.Phrase("41110"))

	return srv.(*httpServer)
}

func TestServer_Serve(t *testing.T) {
	a := assert.New(t, false)

	srv := newTestServer(a, nil)
	a.Equal(srv.State(), web.Stopped)
	router := srv.NewRouter("default", nil)
	router.Get("/mux/test", buildHandler(202))
	router.Get("/m1/test", buildHandler(202))

	router.Get("/m2/test", func(ctx *web.Context) web.Responser {
		srv := ctx.Server()
		a.NotNil(srv)

		ctx.WriteHeader(http.StatusAccepted)
		_, err := ctx.Write([]byte("1234567890"))
		if err != nil {
			println(err)
		}
		return nil
	})

	defer servertest.Run(a, srv)()
	defer srv.Close(0)

	a.PanicString(func() { // 多次调用 srv.Serve
		a.NotError(srv.Serve())
	}, "当前已经处于运行状态")

	servertest.Get(a, "http://localhost:8080/m1/test").Do(nil).Status(http.StatusAccepted)

	servertest.Get(a, "http://localhost:8080/m2/test").Do(nil).Status(http.StatusAccepted)

	servertest.Get(a, "http://localhost:8080/mux/test").Do(nil).Status(http.StatusAccepted)

	// 静态文件
	router.Get("/admin/{path}", web.FileServer(os.DirFS("./testdata"), "path", "index.html"))
	servertest.Get(a, "http://localhost:8080/admin/file1.txt").Do(nil).Status(http.StatusOK)
}

func TestServer_Serve_HTTPS(t *testing.T) {
	a := assert.New(t, false)

	cert, err := tls.LoadX509KeyPair("./testdata/cert.pem", "./testdata/key.pem")
	a.NotError(err).NotNil(cert)
	srv := newTestServer(a, &Options{
		HTTPServer: &http.Server{
			Addr: ":8088",
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		},
	})

	router := srv.NewRouter("default", nil)
	router.Get("/mux/test", buildHandler(202))

	defer servertest.Run(a, srv)()
	defer srv.Close(0)

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
}

func TestServer_Close(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, nil)
	router := srv.NewRouter("def", nil)

	router.Get("/test", buildHandler(202))
	router.Get("/close", func(ctx *web.Context) web.Responser {
		_, err := ctx.Write([]byte("closed"))
		if err != nil {
			ctx.WriteHeader(http.StatusInternalServerError)
		}
		a.Equal(srv.State(), web.Running)
		srv.Close(0)
		srv.Close(0) // 可多次调用
		a.Equal(srv.State(), web.Stopped)
		return nil
	})

	close := 0
	srv.OnClose(func() error {
		close++
		return nil
	})

	defer servertest.Run(a, srv)()
	// defer srv.Close() // 由 /close 关闭，不需要 srv.Close

	servertest.Get(a, "http://localhost:8080/test").Do(nil).Status(http.StatusAccepted)

	// 连接被关闭，返回错误内容
	a.Equal(0, close)
	resp, err := http.Get("http://localhost:8080/close")
	time.Sleep(500 * time.Microsecond) // Handle 中的 Server.Close 是触发关闭服务，这里要等待真正完成
	a.Error(err).Nil(resp).True(close > 0)

	resp, err = http.Get("http://localhost:8080/test")
	a.Error(err).Nil(resp)
}

func TestServer_CloseWithTimeout(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, nil)
	router := srv.NewRouter("def", nil)

	router.Get("/test", buildHandler(202))
	router.Get("/close", func(ctx *web.Context) web.Responser {
		ctx.WriteHeader(http.StatusCreated)
		_, err := ctx.Write([]byte("shutdown with ctx"))
		a.NotError(err)
		srv.Close(300 * time.Millisecond)

		return nil
	})

	defer servertest.Run(a, srv)()
	defer srv.Close(0)

	servertest.Get(a, "http://localhost:8080/test").Do(nil).Status(http.StatusAccepted)

	// 关闭指令可以正常执行
	servertest.Get(a, "http://localhost:8080/close").Do(nil).Status(http.StatusCreated)

	// 未超时，但是拒绝新的链接
	resp, err := http.Get("http://localhost:8080/test")
	a.Error(err).Nil(resp)

	// 已被关闭
	time.Sleep(30 * time.Microsecond)
	resp, err = http.Get("http://localhost:8080/test")
	a.Error(err).Nil(resp)
}

type object struct {
	Name string
	Age  int
}

func TestServer_NewClient(t *testing.T) {
	a := assert.New(t, false)

	s := newTestServer(a, nil)
	defer servertest.Run(a, s)()
	defer s.Close(500 * time.Millisecond)

	s.NewRouter("default", nil).Get("/get", func(ctx *web.Context) web.Responser {
		return web.OK(&object{Name: "name"})
	}).Post("/post", func(ctx *web.Context) web.Responser {
		obj := &object{}
		if resp := ctx.Read(true, obj, web.ProblemBadRequest); resp != nil {
			return resp
		}
		if obj.Name != "name" {
			return ctx.Problem(web.ProblemBadRequest).WithExtensions(&object{Name: "name"})
		}
		return web.OK(obj)
	})

	c := s.NewClient(nil, "http://localhost:8080", "application/json")
	a.NotNil(c)

	resp := &object{}
	p := &web.RFC7807{}
	a.NotError(c.Get("/get", resp, p))
	a.Zero(p).Equal(resp, &object{Name: "name"})

	resp = &object{}
	p = &web.RFC7807{}
	a.NotError(c.Delete("/get", resp, p))
	a.Zero(resp).Equal(p.Type, web.ProblemMethodNotAllowed)

	resp = &object{}
	p = &web.RFC7807{Extensions: &object{}}
	a.NotError(c.Post("/post", nil, resp, p))
	a.Zero(resp).
		Equal(p.Type, web.ProblemBadRequest).
		Equal(p.Extensions, &object{Name: "name"})

	resp = &object{}
	p = &web.RFC7807{}
	a.NotError(c.Post("/post", &object{Age: 1, Name: "name"}, resp, p))
	a.Zero(p).Equal(resp, &object{Age: 1, Name: "name"})

	resp = &object{}
	p = &web.RFC7807{}
	a.NotError(c.Patch("/not-exists", nil, resp, p))
	a.Zero(resp).Equal(p.Type, web.ProblemNotFound)
}
