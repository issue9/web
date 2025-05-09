// SPDX-FileCopyrightText: 2018-2025 caixw
//
// SPDX-License-Identifier: MIT

package server

import (
	"crypto/tls"
	sj "encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
	"github.com/issue9/cache"
	"github.com/issue9/cache/caches/memory"
	"github.com/issue9/logs/v7"
	"github.com/issue9/mux/v9"
	"github.com/issue9/mux/v9/header"
	"golang.org/x/text/language"

	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/json"
	"github.com/issue9/web/mimetype/xml"
	"github.com/issue9/web/selector"
	"github.com/issue9/web/server/registry"
	"github.com/issue9/web/server/servertest"
)

var (
	_ web.Server = &httpServer{}
	_ web.Server = &gateway{}
	_ web.Server = &service{}
)

func buildHandler(code int) web.HandlerFunc {
	return func(ctx *web.Context) web.Responser {
		return web.ResponserFunc(func(ctx *web.Context) { ctx.Render(code, code) })
	}
}

func TestNew(t *testing.T) {
	a := assert.New(t, false)

	srv, err := NewHTTP("app", "0.1.0", nil)
	a.NotError(err).NotNil(srv).
		False(srv.Uptime().IsZero()).
		NotNil(srv.Cache()).
		Equal(srv.Location(), time.Local)

	s, ok := srv.(*httpServer)
	a.True(ok).Equal(s.hs.Handler, s).
		Equal(s.hs.Addr, "")

	d, ok := srv.Cache().(cache.Driver)
	a.True(ok).
		NotNil(d).
		NotNil(d.Driver()).
		True(srv.CanCompress())
	srv.SetCompress(false)
	a.False(srv.CanCompress())
}

func newOptions(o *Options) *Options {
	if o == nil {
		o = &Options{HTTPServer: &http.Server{Addr: ":8080"}, Language: language.English} // 指定不存在的语言
	}
	if o.Logs == nil { // 默认重定向到 os.Stderr
		o.Logs = logs.New(logs.NewTermHandler(os.Stderr, nil), logs.WithLocation(true), logs.WithCreated(logs.NanoLayout), logs.WithLevels(logs.AllLevels()...))
	}
	if o.Codec == nil {
		o.Codec = web.NewCodec().
			AddMimetype(json.Mimetype, json.Marshal, json.Unmarshal, json.ProblemMimetype, true, true).
			AddMimetype(xml.Mimetype, xml.Marshal, xml.Unmarshal, "", true, true)
	}

	return o
}

func newTestServer(a *assert.Assertion, o *Options) web.Server {
	srv, err := NewHTTP("app", "0.1.0", newOptions(o))
	a.NotError(err).NotNil(srv).
		Equal(srv.ID(), "app").Equal(srv.Version(), "0.1.0")

	l := srv.Locale()
	a.NotError(l.SetString(language.Und, "lang", "und")).
		NotError(l.SetString(language.SimplifiedChinese, "lang", "hans")).
		NotError(l.SetString(language.TraditionalChinese, "lang", "hant"))

	srv.Problems().Add(411, &web.LocaleProblem{ID: "41110", Title: web.Phrase("lang"), Detail: web.Phrase("41110")})

	return srv
}

func TestHTTPServer_Serve(t *testing.T) {
	a := assert.New(t, false)

	srv := newTestServer(a, nil)
	a.Equal(srv.State(), web.Stopped)
	router := srv.Routers().New("default", nil)
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

	router.Get("/m2/103", func(ctx *web.Context) web.Responser {
		ctx.Header().Set("h1", "v1")
		ctx.WriteHeader(http.StatusEarlyHints)
		return web.OK("123")
	})

	defer servertest.Run(a, srv)()
	defer srv.Close(0)

	a.PanicString(func() { // 多次调用 srv.Serve
		a.NotError(srv.Serve())
	}, "当前已经处于运行状态")

	servertest.Get(a, "http://localhost:8080/m1/test").Do(nil).Status(http.StatusAccepted)
	servertest.Get(a, "http://localhost:8080/m2/test").Do(nil).Status(http.StatusAccepted)
	servertest.Get(a, "http://localhost:8080/m2/103").Do(nil).Status(http.StatusOK)
	servertest.Get(a, "http://localhost:8080/mux/test").Do(nil).Status(http.StatusAccepted)
}

func TestHTTPServer_Serve_HTTPS(t *testing.T) {
	a := assert.New(t, false)

	cert, err := tls.LoadX509KeyPair("./config/testdata/cert.pem", "./config/testdata/key.pem")
	a.NotError(err).NotNil(cert)
	srv := newTestServer(a, &Options{
		HTTPServer: &http.Server{
			Addr: ":8088",
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{cert},
			},
		},
	})

	router := srv.Routers().New("default", nil)
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

func TestHTTPServer_Close(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, nil)
	router := srv.Routers().New("def", nil)

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

	c := 0
	srv.OnClose(func() error {
		c++
		return nil
	})

	defer servertest.Run(a, srv)()
	// defer srv.Close() // 由 /close 关闭，不需要 srv.Close

	servertest.Get(a, "http://localhost:8080/test").Do(nil).Status(http.StatusAccepted)

	// 尝试关闭连接
	a.Equal(0, c)
	resp, err := http.Get("http://localhost:8080/close")
	a.Wait(500 * time.Microsecond). // Handle 中的 Server.Close 是触发关闭服务，这里要等待真正完成
					NotError(err).NotNil(resp).True(c > 0)

	// 连接被关闭，返回错误内容
	resp, err = http.Get("http://localhost:8080/close")
	a.Error(err).Nil(resp)

	// 连接被关闭，返回错误内容
	resp, err = http.Get("http://localhost:8080/test")
	a.Error(err).Nil(resp)
}

func TestHTTPServer_CloseWithTimeout(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, nil)
	router := srv.Routers().New("def", nil)

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
	a.Error(err).Nil(resp).Wait(30 * time.Microsecond)

	// 已被关闭
	resp, err = http.Get("http://localhost:8080/test")
	a.Error(err).Nil(resp)
}

type object struct {
	Name string
	Age  int
}

func TestHTTPServer_NewClient(t *testing.T) {
	a := assert.New(t, false)

	s := newTestServer(a, nil)
	defer servertest.Run(a, s)()
	defer s.Close(500 * time.Millisecond)

	s.Routers().New("default", nil).Get("/get", func(ctx *web.Context) web.Responser {
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

	sel := selector.NewRoundRobin(false, 1)
	sel.Update(selector.NewPeer("http://localhost:8080"))
	c := s.NewClient(nil, sel, header.JSON, sj.Marshal)
	a.NotNil(c)

	resp := &object{}
	a.NotError(c.Get("/get", resp, nil)).Equal(resp, &object{Name: "name"})

	resp = &object{}
	err := c.Delete("/get", resp, nil)
	a.Error(err).Zero(resp)
	p, ok := err.(*web.Problem)
	a.True(ok).NotNil(p).
		Equal(p.Type, web.ProblemMethodNotAllowed).
		Equal(p.Status, http.StatusMethodNotAllowed)

	resp = &object{}
	pb := func() *web.Problem { return &web.Problem{Extensions: &object{}} }
	err = c.Post("/post", nil, resp, pb)
	a.Error(err).Zero(resp)
	p, ok = err.(*web.Problem)
	a.True(ok).NotNil(p).
		Equal(p.Type, web.ProblemBadRequest).
		Equal(p.Extensions, &object{Name: "name"})

	resp = &object{}
	a.NotError(c.Post("/post", &object{Age: 1, Name: "name"}, resp, nil)).
		Equal(resp, &object{Age: 1, Name: "name"})

	resp = &object{}
	err = c.Patch("/not-exists", nil, resp, nil)
	a.Error(err).Zero(resp)
	p, ok = err.(*web.Problem)
	a.True(ok).NotNil(p).
		Equal(p.Type, web.ProblemNotFound)
}

func TestHTTPServer_NewContext(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, nil)
	router := srv.Routers().New("def", nil)

	defer servertest.Run(a, srv)()
	defer srv.Close(0)

	// 正常，指定 Accept-Language，采用默认的 accept
	router.Get("/path", func(ctx *web.Context) web.Responser {
		a.NotNil(ctx).NotEmpty(ctx.ID())
		a.Equal(ctx.Mimetype(false), header.JSON).
			Equal(ctx.Charset(), header.UTF8).
			Equal(ctx.LanguageTag(), language.SimplifiedChinese).
			NotNil(ctx.LocalePrinter())
		return nil
	})
	servertest.Get(a, "http://localhost:8080/path").
		Header(header.AcceptLanguage, "cmn-hans").
		Header(header.Accept, header.JSON).
		Do(nil).
		Success()
}

func TestNewService(t *testing.T) {
	a := assert.New(t, false)

	// Registry 和 Peer 是空的
	srv, err := NewService("app", "0.1.0", newOptions(nil))
	a.Error(err).Nil(srv)

	c := memory.New()
	reg := registry.NewCache(web.NewCache("reg:", c), registry.NewRandomStrategy(), time.Second)

	srv = newService(a, "app", ":8080", reg, c)
	a.Equal(srv.ID(), "app").Equal(srv.Version(), "0.1.0")

	srv.Routers().New("default", nil).Get("/mux/test", buildHandler(202))

	defer servertest.Run(a, srv)()
	defer srv.Close(0)

	servertest.Get(a, "http://localhost:8080/mux/test").Do(nil).Status(202)
}

func newService(a *assert.Assertion, name, addr string, reg registry.Registry, c cache.Driver) web.Server {
	srv, err := NewService(name, "0.1.0", newOptions(&Options{
		Cache:      c,
		HTTPServer: &http.Server{Addr: addr},
		Registry:   reg,
		Peer:       selector.NewPeer("http://localhost" + addr),
	}))
	a.NotError(err).NotNil(srv)

	return srv
}

func TestNewGateway(t *testing.T) {
	a := assert.New(t, false)
	c := memory.New() // 默认的缓存系统用的是内存类型的，保证引用同一个。

	reg := registry.NewCache(web.NewCache("reg:", c), registry.NewRandomStrategy(), time.Second)

	// s1
	s1 := newService(a, "s1", ":8081", reg, c)
	defer servertest.Run(a, s1)()
	defer s1.Close(0)
	s1.Routers().New("default", nil).Get("/mux/test", buildHandler(201))

	// s2
	s2 := newService(a, "s2", ":8082", reg, c)
	defer servertest.Run(a, s2)()
	defer s2.Close(0)
	s2.Routers().New("default", nil).Get("/mux/test", buildHandler(202))

	// Registry 和 Mapper 是空的
	g, err := NewGateway("app", "0.1.0", newOptions(nil))
	a.Equal(err.(*web.FieldError).Field, "o.Mapper").Nil(g)

	g, err = NewGateway("app", "0.1.0", newOptions(&Options{
		Cache:      c,
		HTTPServer: &http.Server{Addr: ":8080"},
		Registry:   reg,
		Mapper: map[string]mux.Matcher{
			"s1": mux.NewPathVersion("", "/s1"),
			"s2": mux.NewPathVersion("", "/s2"),
		},
	}))
	a.NotError(err).NotNil(g)
	a.Equal(g.ID(), "app").Equal(g.Version(), "0.1.0")
	g.Routers().New("default", nil).Get("/mux/test", buildHandler(203))

	defer servertest.Run(a, g)()
	defer g.Close(0)

	servertest.Get(a, "http://localhost:8080/s1/mux/test").Do(nil).Status(201)
	servertest.Get(a, "http://localhost:8080/s2/mux/test").Do(nil).Status(202)
	servertest.Get(a, "http://localhost:8080/s3/mux/test").Do(nil).Status(http.StatusNotFound)
	servertest.Get(a, "http://localhost:8080/mux/test").Do(nil).Status(203)
}
