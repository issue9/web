// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/issue9/assert/v6"
	"github.com/issue9/assert/v6/rest"
	"github.com/issue9/cache/drivers/memory"
	"github.com/issue9/config"
	"github.com/issue9/logs/v7"
	"github.com/issue9/mux/v10/header"
	"github.com/issue9/unique/v2"
	"golang.org/x/text/language"

	"github.com/issue9/web/internal/servertest"
	"github.com/issue9/web/selector"
)

var _ context.Context = &Server{}

var portNum atomic.Int64

func init() { portNum.Store(8080) }

type testServer struct {
	*Server
	logBuf *bytes.Buffer
	st     *servertest.Server
}

func buildHandler(code int) HandlerFunc {
	return func(ctx *Context) Responser {
		return ResponserFunc(func(ctx *Context) { ctx.Render(code, code) })
	}
}

// 创建一个自定义了日志系统的服务
//
// sec 是否使用 https
func newTestServer(a *assert.Assertion, sec bool) *testServer {
	logBuf := new(bytes.Buffer)
	log := logs.New(
		logs.NewTextHandler(logBuf, os.Stderr),
		logs.WithCreated(logs.NanoLayout),
		logs.WithLocation(true),
		logs.WithLevels(logs.AllLevels()...),
	)
	a.NotNil(log)

	conf := config.Dir(nil, "./")

	srv := &testServer{
		logBuf: logBuf,
		st:     servertest.NewServer(a, sec),
	}

	cc := memory.New()
	u := unique.NewNumber(100)
	hs := &http.Server{Addr: srv.st.Port}
	if sec {
		cert, err := tls.LoadX509KeyPair("./server/config/testdata/cert.pem", "./server/config/testdata/key.pem")
		a.NotError(err).NotNil(cert)
		hs.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}
	startup := NewHTTPServerStartup(hs)

	srv.Server = NewServer("test", "1.0.0", conf, startup, time.Local, language.SimplifiedChinese, log, u.String, cc, newCodec(a), header.XRequestID, "")
	srv.Services().Add(Phrase("unique"), u)

	a.NotError(srv.Locale().SetString(language.SimplifiedChinese, "lang", "cn")).
		NotError(srv.Locale().SetString(language.TraditionalChinese, "lang", "tw"))

	srv.Problems().Add(411, &LocaleProblem{ID: "41110", Title: Phrase("41110 title"), Detail: Phrase("41110 detail")})

	return srv
}

func (s *testServer) NewRequest(method, path string) *rest.Request {
	return rest.NewRequest(s.st.A, method, s.st.URL+path).Client(s.st.Client)
}

func (s *testServer) Get(path string) *rest.Request { return s.NewRequest(http.MethodGet, path) }

func (s *testServer) Patch(path string) *rest.Request { return s.NewRequest(http.MethodPatch, path) }

func (s *testServer) Run() func() { return servertest.Run(s.st.A, s) }

func onExitContext(ctx *Context, status int) {
	ctx.Header().Set("exit-status", strconv.Itoa(status))
}

func TestOnExitContextFunc(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a, false)
	s.OnExitContext(onExitContext)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/path", nil)
	ctx := s.NewContext(w, r, nil)
	ctx.WriteHeader(http.StatusAccepted)
	ctx.freeContext()
	a.Equal(ctx.Header().Get("exit-status"), "202")
}

func TestHTTPServer_Serve(t *testing.T) {
	a := assert.New(t, false)

	srv := newTestServer(a, false)
	a.Equal(srv.State(), Stopped)
	router := srv.Routers().New("default", nil)
	router.Get("/mux/test", buildHandler(202))
	router.Get("/m1/test", buildHandler(202))

	router.Get("/m2/test", func(ctx *Context) Responser {
		srv := ctx.Server()
		a.NotNil(srv)

		ctx.WriteHeader(http.StatusAccepted)
		_, err := ctx.Write([]byte("1234567890"))
		if err != nil {
			println(err)
		}
		return nil
	})

	router.Get("/m2/103", func(ctx *Context) Responser {
		ctx.Header().Set("h1", "v1")
		ctx.WriteHeader(http.StatusEarlyHints)
		return OK("123")
	})

	defer srv.Run()()
	defer srv.Close(0)

	a.PanicString(func() { // 多次调用 srv.Serve
		a.NotError(srv.Serve())
	}, "当前已经处于运行状态")

	srv.Get("/m1/test").Do().Status(http.StatusAccepted)
	srv.Get("/m2/test").Do().Status(http.StatusAccepted)
	srv.Get("/m2/103").Do().Status(http.StatusOK)
	srv.Get("/mux/test").Do().Status(http.StatusAccepted)
}

func TestHTTPServer_Serve_HTTPS(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, true)

	router := srv.Routers().New("default", nil)
	router.Get("/mux/test", buildHandler(202))

	defer srv.Run()()
	defer srv.Close(0)

	srv.Get("/mux/test").Do().Status(http.StatusAccepted)

	// 无效的 http 请求

	client := srv.st.Client
	resp, err := client.Get("http://localhost" + srv.st.Port + "/mux/test")
	a.NotError(err).Equal(resp.StatusCode, http.StatusBadRequest)

	resp, err = client.Get("http://localhost" + srv.st.Port + "/mux")
	a.NotError(err).Equal(resp.StatusCode, http.StatusBadRequest)
}

func TestHTTPServer_Close(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, false)
	router := srv.Routers().New("def", nil)

	router.Get("/test", buildHandler(202))
	router.Get("/close", func(ctx *Context) Responser {
		_, err := ctx.Write([]byte("closed"))
		if err != nil {
			ctx.WriteHeader(http.StatusInternalServerError)
		}
		a.Equal(srv.State(), Running)
		srv.Close(0)
		srv.Close(0) // 可多次调用
		a.Equal(srv.State(), Stopped)
		return nil
	})

	c := 0
	srv.OnClose(func() error {
		c++
		return nil
	})

	defer srv.Run()()
	// defer srv.Close() // 由 /close 关闭，不需要 srv.Close

	srv.Get("/test").Do().Status(http.StatusAccepted)

	// 尝试关闭连接
	a.Equal(0, c)
	srv.Get("/close").Do()
	time.Sleep(500 * time.Microsecond) // Handle 中的 Server.Close 是触发关闭服务，这里要等待真正完成
	a.True(c > 0)

	// 连接被关闭，返回错误内容
	resp, err := http.Get(srv.st.URL + "/close")
	a.Error(err).Nil(resp)

	// 连接被关闭，返回错误内容
	resp, err = http.Get(srv.st.URL + "/close")
	a.Error(err).Nil(resp)
}

func TestHTTPServer_CloseWithTimeout(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, false)
	router := srv.Routers().New("def", nil)

	router.Get("/test", buildHandler(202))
	router.Get("/close", func(ctx *Context) Responser {
		ctx.WriteHeader(http.StatusCreated)
		_, err := ctx.Write([]byte("shutdown with ctx"))
		a.NotError(err)
		srv.Close(300 * time.Millisecond)

		return nil
	})

	defer srv.Run()()
	defer srv.Close(0)

	srv.Get("/test").Do().Status(http.StatusAccepted)

	// 关闭指令可以正常执行
	srv.Get("/close").Do().Status(http.StatusCreated)

	// 未超时，但是拒绝新的链接
	resp, err := http.Get(srv.st.URL + "/test")
	a.Error(err).Nil(resp)
	time.Sleep(30 * time.Microsecond)

	// 已被关闭
	resp, err = http.Get(srv.st.URL + "/test")
	a.Error(err).Nil(resp)
}

func TestHTTPServer_NewClient(t *testing.T) {
	a := assert.New(t, false)

	s := newTestServer(a, false)
	defer s.Run()()
	defer s.Close(500 * time.Millisecond)

	s.Routers().New("default", nil).Get("/get", func(ctx *Context) Responser {
		return OK(&object{Name: "name"})
	}).Post("/post", func(ctx *Context) Responser {
		obj := &object{}
		if resp := ctx.Read(true, obj, ProblemBadRequest); resp != nil {
			return resp
		}
		if obj.Name != "name" {
			return ctx.Problem(ProblemBadRequest).WithExtensions(&object{Name: "name"})
		}
		return OK(obj)
	})

	sel := selector.NewRoundRobin(false, 1)
	sel.Update(selector.NewPeer(s.st.URL))
	c := s.NewClient(s.st.Client, sel, header.JSON, func(a any) ([]byte, error) { return json.Marshal(a) })
	a.NotNil(c)

	resp, p, err := c.Get[object, any]("/get")
	a.NotError(err).
		Equal(resp, &object{Name: "name"}).
		Nil(p)

	resp, p, err = c.Delete[object, any]("/get")
	a.NotError(err).
		Zero(resp).
		NotNil(p).
		Equal(p.Type, ProblemMethodNotAllowed).
		Equal(p.Status, http.StatusMethodNotAllowed)

	resp, p, err = c.Post[object, *object]("/post", nil)
	a.NotError(err).
		Zero(resp).
		NotNil(p).
		Equal(p.Type, ProblemBadRequest).
		Equal(p.Extensions, &object{Name: "name"})

	resp, p, err = c.Post[object, any]("/post", &object{Age: 1, Name: "name"})
	a.NotError(err).
		Nil(p).
		Equal(resp, &object{Age: 1, Name: "name"})

	resp, p, err = c.Patch[object, any]("/not-exists", nil)
	a.NotError(err).
		Zero(resp).
		NotNil(p).
		Equal(p.Type, ProblemNotFound)
}

func TestHTTPServer_NewContext(t *testing.T) {
	a := assert.New(t, false)
	srv := newTestServer(a, false)
	router := srv.Routers().New("def", nil)

	defer srv.Run()()
	defer srv.Close(0)

	// 正常，指定 Accept-Language，采用默认的 accept
	router.Get("/path", func(ctx *Context) Responser {
		a.NotNil(ctx).NotEmpty(ctx.ID())
		a.Equal(ctx.Mimetype(false), header.JSON).
			Equal(ctx.Charset(), header.UTF8).
			Equal(ctx.LanguageTag(), language.SimplifiedChinese).
			NotNil(ctx.LocalePrinter())
		return nil
	})
	srv.Get("/path").
		Header(header.AcceptLanguage, "cmn-hans").
		Header(header.Accept, header.JSON).
		Do().
		Success()
}
