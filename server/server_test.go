// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
	"github.com/issue9/config"
	"github.com/issue9/logs/v2"
	"github.com/issue9/middleware/compress"
	"gopkg.in/yaml.v2"

	wctx "github.com/issue9/web/context"
	"github.com/issue9/web/context/mimetype"
	"github.com/issue9/web/context/mimetype/gob"
	"github.com/issue9/web/internal/webconfig"
)

var f202 = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	_, err := w.Write([]byte("1234567890"))
	if err != nil {
		println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// 声明一个 Server 实例
func newServer(a *assert.Assertion) *Server {
	var configUnmarshals = map[string]config.UnmarshalFunc{
		".yaml": yaml.Unmarshal,
		".yml":  yaml.Unmarshal,
		".xml":  xml.Unmarshal,
		".json": json.Unmarshal,
	}

	mgr, err := config.NewManager("./testdata")
	a.NotError(err).NotNil(mgr)
	for k, v := range configUnmarshals {
		a.NotError(mgr.AddUnmarshal(v, k))
	}

	webconf := &webconfig.WebConfig{}
	a.NotError(mgr.LoadFile("web.yaml", webconf))

	app, err := New(webconf, logs.New(), wctx.DefaultResultBuilder)
	a.NotError(err).NotNil(app)

	a.NotError(app.AddCompresses(map[string]compress.WriterFunc{
		"gzip":    compress.NewGzip,
		"deflate": compress.NewDeflate,
	}))

	a.NotError(app.Builder().AddMarshals(map[string]mimetype.MarshalFunc{
		"application/json":       json.Marshal,
		"application/xml":        xml.Marshal,
		mimetype.DefaultMimetype: gob.Marshal,
	}))

	a.NotError(app.Builder().AddUnmarshals(map[string]mimetype.UnmarshalFunc{
		"application/json":       json.Unmarshal,
		"application/xml":        xml.Unmarshal,
		mimetype.DefaultMimetype: gob.Unmarshal,
	}))

	// 以下内容由配置文件决定
	a.True(app.IsDebug()).
		NotNil(app.webConfig.Compress).
		NotEmpty(app.compresses)

	a.NotNil(app.Builder()).Equal(app.Builder(), app.builder)
	a.NotNil(app.Server.Handler)
	a.NotNil(app.ErrorHandlers())
	a.NotNil(app.Logs())
	a.NotNil(app.Mux())
	a.Equal(app.Mux(), app.router.Mux())

	return app
}

func TestApp_URL(t *testing.T) {
	a := assert.New(t)
	app := newServer(a)

	a.Equal(app.URL("/abc"), "http://localhost:8082/abc")
	a.Equal(app.URL("abc/def"), "http://localhost:8082/abc/def")
	a.Equal(app.URL(""), "http://localhost:8082")
}

func TestApp_Path(t *testing.T) {
	a := assert.New(t)
	app := newServer(a)

	a.Equal(app.Path("/abc"), "/abc")
	a.Equal(app.Path("abc/def"), "/abc/def")
	a.Equal(app.Path(""), "")
}

func TestApp_Run(t *testing.T) {
	a := assert.New(t)
	app := newServer(a)
	exit := make(chan bool, 1)
	err := app.ErrorHandlers().Add(func(w http.ResponseWriter, status int) {
		w.WriteHeader(status)
		_, err := w.Write([]byte("error handler test"))
		a.NotError(err)
	}, http.StatusNotFound)
	a.NotError(err)

	app.Mux().GetFunc("/m1/test", f202)
	app.Mux().GetFunc("/m2/test", f202)
	app.Mux().GetFunc("/mux/test", f202)

	go func() {
		err := app.Run()
		a.ErrorType(err, http.ErrServerClosed, "assert.ErrorType 错误，%v", err)
		exit <- true
	}()
	time.Sleep(500 * time.Microsecond) // 等待 go func() 完成

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8082/m1/test").
		Do().
		Status(http.StatusAccepted)

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8082/m2/test").
		Do().
		Status(http.StatusAccepted)

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8082/mux/test").
		Do().
		Status(http.StatusAccepted)

	// not found
	// 返回 ErrorHandler 内容
	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8082/mux/not-exists.txt").
		Do().
		Status(http.StatusNotFound).
		StringBody("error handler test")

	// static 中定义的静态文件
	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8082/client/file1.txt").
		Do().
		Status(http.StatusOK)

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8082/client/dir/file2.txt").
		Do().
		Status(http.StatusOK)

	// 不存在的文件，测试 internal/fileserver 是否启作用
	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8082/client/dir/not-exists.txt").
		Do().
		Status(http.StatusNotFound).
		StringBody("error handler test")

	a.NotError(app.Close())
	<-exit
}

func TestApp_Close(t *testing.T) {
	a := assert.New(t)
	app := newServer(a)
	exit := make(chan bool, 1)

	app.Mux().GetFunc("/test", f202)
	app.Mux().GetFunc("/close", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("closed"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		a.NotError(app.Close())
	})

	go func() {
		err := app.Run()
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
		exit <- true
	}()

	// 等待 app.Serve() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(500 * time.Microsecond)

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8082/test").
		Do().
		Status(http.StatusAccepted)

	// 连接被关闭，返回错误内容
	resp, err := http.Get("http://localhost:8082/close")
	a.Error(err).Nil(resp)

	resp, err = http.Get("http://localhost:8082/test")
	a.Error(err).Nil(resp)

	<-exit
}

func TestApp_Shutdown(t *testing.T) {
	a := assert.New(t)
	app := newServer(a)
	exit := make(chan bool, 1)

	app.Mux().GetFunc("/test", f202)
	app.Mux().GetFunc("/close", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, err := w.Write([]byte("shutdown with ctx"))
		a.NotError(err)
		ctx, c := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer c()
		app.Shutdown(ctx)
	})

	go func() {
		err := app.Run()
		a.Error(err).ErrorType(err, http.ErrServerClosed, "错误信息为:%v", err)
		exit <- true
	}()

	// 等待 app.Serve() 启动完毕，不同机器可能需要的时间会不同
	time.Sleep(500 * time.Microsecond)

	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8082/test").
		Do().
		Status(http.StatusAccepted)

	// 关闭指令可以正常执行
	rest.NewRequest(a, nil, http.MethodGet, "http://localhost:8082/close").
		Do().
		Status(http.StatusCreated)

	// 未超时，但是拒绝新的链接
	resp, err := http.Get("http://localhost:8082/test")
	a.Error(err).Nil(resp)

	// 已被关闭
	time.Sleep(30 * time.Microsecond)
	resp, err = http.Get("http://localhost:8082/test")
	a.Error(err).Nil(resp)

	<-exit
}
