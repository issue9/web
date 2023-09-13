// SPDX-License-Identifier: MIT

package web

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"io/fs"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"golang.org/x/text/language"

	"github.com/issue9/web/logs"
	"github.com/issue9/web/servertest"
)

var (
	_ fs.FS             = &Server{}
	_ servertest.Server = &Server{}
)

func TestServer_Vars(t *testing.T) {
	a := assert.New(t, false)
	srv, err := NewServer("app", "1.0.0", nil)
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

func buildHandler(code int) HandlerFunc {
	return func(ctx *Context) Responser {
		return ResponserFunc(func(ctx *Context) *Problem {
			ctx.Render(code, code)
			return nil
		})
	}
}

func newTestServer(a *assert.Assertion, o *Options) *Server {
	if o == nil {
		o = &Options{HTTPServer: &http.Server{Addr: ":8080"}, Locale: &Locale{Language: language.English}} // 指定不存在的语言
	}
	if o.Logs == nil { // 默认重定向到 os.Stderr
		o.Logs = &logs.Options{
			Handler: logs.NewTermHandler(logs.NanoLayout, os.Stderr, nil),
			Caller:  true,
			Created: true,
			Levels:  logs.AllLevels(),
		}
	}
	if o.Encodings == nil {
		o.Encodings = []*Encoding{
			{Name: "gzip", Builder: GZipWriter(8)},
			{Name: "deflate", Builder: DeflateWriter(8)},
		}
	}
	if o.Mimetypes == nil {
		o.Mimetypes = []*Mimetype{
			{Type: "application/json", Marshal: marshalJSON, Unmarshal: json.Unmarshal, ProblemType: "application/problem+json"},
			{Type: "application/xml", Marshal: marshalXML, Unmarshal: xml.Unmarshal, ProblemType: ""},
			{Type: "application/test", Marshal: marshalTest, Unmarshal: unmarshalTest, ProblemType: ""},
			{Type: "nil", Marshal: nil, Unmarshal: nil, ProblemType: ""},
		}
	}

	srv, err := NewServer("app", "0.1.0", o)
	a.NotError(err).NotNil(srv)
	a.Equal(srv.Name(), "app").Equal(srv.Version(), "0.1.0")

	// locale
	b := srv.CatalogBuilder()
	a.NotError(b.SetString(language.Und, "lang", "und"))
	a.NotError(b.SetString(language.SimplifiedChinese, "lang", "hans"))
	a.NotError(b.SetString(language.TraditionalChinese, "lang", "hant"))

	srv.AddProblem("41110", 411, Phrase("lang"), Phrase("41110"))

	return srv
}

func TestNewServer(t *testing.T) {
	a := assert.New(t, false)

	srv, err := NewServer("app", "0.1.0", nil)
	a.NotError(err).NotNil(srv).
		False(srv.Uptime().IsZero()).
		NotNil(srv.Cache()).
		Equal(srv.Location(), time.Local).
		Equal(srv.httpServer.Handler, srv.routers).
		Equal(srv.httpServer.Addr, "")
}

func TestServer_Serve(t *testing.T) {
	a := assert.New(t, false)

	srv := newTestServer(a, nil)
	a.Equal(srv.State(), Stopped)
	router := srv.NewRouter("default", nil)
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

	defer servertest.Run(a, srv)()
	defer srv.Close(0)

	a.PanicString(func() { // 多次调用 srv.Serve
		a.NotError(srv.Serve())
	}, "当前已经处于运行状态")

	servertest.Get(a, "http://localhost:8080/m1/test").Do(nil).Status(http.StatusAccepted)

	servertest.Get(a, "http://localhost:8080/m2/test").Do(nil).Status(http.StatusAccepted)

	servertest.Get(a, "http://localhost:8080/mux/test").Do(nil).Status(http.StatusAccepted)

	// 静态文件
	router.Get("/admin/{path}", srv.FileServer(os.DirFS("./testdata"), "path", "index.html"))
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
	router.Get("/close", func(ctx *Context) Responser {
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

// 检测 204 是否存在 http: request method or response status code does not allow body
func TestContext_NoContent(t *testing.T) {
	a := assert.New(t, false)
	buf := new(bytes.Buffer)
	o := &Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Logs:       &logs.Options{Handler: logs.NewTextHandler("15:04:05", buf)},
	}
	s := newTestServer(a, o)

	s.NewRouter("def", nil).Get("/204", func(ctx *Context) Responser {
		return ResponserFunc(func(ctx *Context) *Problem {
			ctx.WriteHeader(http.StatusNoContent)
			return nil
		})
	})

	defer servertest.Run(a, s)()

	servertest.Get(a, "http://localhost:8080/204").
		Header("Accept-Encoding", "gzip"). // 服务端不应该构建压缩对象
		Header("Accept", "application/json;charset=gbk").
		Do(nil).
		Status(http.StatusNoContent)

	s.Close(0)

	a.NotContains(buf.String(), "request method or response status code does not allow body")
}
