// SPDX-License-Identifier: MIT

package web

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"

	context2 "github.com/issue9/web/context"
	"github.com/issue9/web/context/mimetype"
	"github.com/issue9/web/context/mimetype/gob"
	"github.com/issue9/web/context/mimetype/mimetypetest"
	"github.com/issue9/web/internal/webconfig"
)

func initApp(a *assert.Assertion) {
	defaultServer = nil
	a.NotError(Classic("./testdata", context2.DefaultResultBuilder))
	a.NotNil(defaultServer)
	a.Equal(defaultServer, Server())

	err := Builder().AddMarshals(map[string]mimetype.MarshalFunc{
		"application/xml":        xml.Marshal,
		mimetype.DefaultMimetype: gob.Marshal,
		mimetypetest.Mimetype:    mimetypetest.TextMarshal,
	})
	a.NotError(err)

	err = Builder().AddUnmarshals(map[string]mimetype.UnmarshalFunc{
		"application/xml":        xml.Unmarshal,
		mimetype.DefaultMimetype: gob.Unmarshal,
		mimetypetest.Mimetype:    mimetypetest.TextUnmarshal,
	})
	a.NotError(err)

	a.NotNil(ErrorHandlers())
	a.Equal(1, len(Services())) // 默认有 scheduled 的服务在运行
}

func TestClassic(t *testing.T) {
	a := assert.New(t)
	initApp(a)

	a.Panic(func() {
		a.NotError(Classic("./testdata", context2.DefaultResultBuilder))
	})

	a.True(IsDebug())
	a.Equal(URL("/test/abc.png"), "http://localhost:8082/test/abc.png")
	a.Equal(Path("/test/abc.png"), "/test/abc.png")
}

func TestSchedulers(t *testing.T) {
	a := assert.New(t)
	initApp(a)

	a.Empty(Schedulers())
	Scheduled().At("test", func(time.Time) error { return nil }, "2001-01-02 17:18:19", false)
	a.Equal(1, len(Schedulers()))
}

func TestMessages(t *testing.T) {
	a := assert.New(t)
	initApp(a)

	a.Empty(Messages(nil))

	AddMessages(http.StatusNotImplemented, map[int]string{
		50010: "50010",
		50011: "50011",
	})

	a.Equal(2, len(Messages(nil)))
}

func TestTime(t *testing.T) {
	a := assert.New(t)
	initApp(a)

	a.Equal(Location(), time.UTC)
	a.Equal(Now().Location(), Location())
	a.Equal(Now().Unix(), time.Now().Unix())
}

func TestModules(t *testing.T) {
	a := assert.New(t)
	initApp(a)
	exit := make(chan bool, 1)

	m1 := NewModule("m1", "m1 desc", "m2")
	m1.AddInit(func() error {
		println("m1")
		return nil
	}, "init")
	m1.PostFunc("/post/m1", func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(w, r)
		ctx.Render(http.StatusCreated, "m1", nil)
	})

	m2 := NewModule("m2", "m2 desc")
	a.NotNil(m2)

	a.NotError(InitModules(""))
	a.Equal(2, len(Modules())) //  m1,m2

	go func() {
		err := Serve()
		if err != http.ErrServerClosed {
			a.NotError(err)
		}
		exit <- true
	}()
	time.Sleep(500 * time.Microsecond)

	rest.NewRequest(a, nil, http.MethodPost, URL("/post")).
		Header("Accept", "application/json").
		Do().
		Status(http.StatusNotFound)

	rest.NewRequest(a, nil, http.MethodPost, URL("/post/m1")).
		Header("Accept", "application/json").
		Do().
		Status(http.StatusCreated).
		StringBody("\"m1\"") // json 格式的字符串

	ctx, c := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer c()
	a.NotError(Shutdown(ctx))

	<-exit
}

func TestFile(t *testing.T) {
	a := assert.New(t)
	initApp(a)

	path, err := filepath.Abs("./testdata/web.yaml")
	a.NotError(err)

	a.Equal(File("web.yaml"), path)

	conf1 := &webconfig.WebConfig{}
	a.NotError(LoadFile("web.yaml", conf1))
	a.NotNil(conf1)

	file, err := os.Open(path)
	a.NotError(err).NotNil(file)
	conf2 := &webconfig.WebConfig{}
	a.NotError(Load(file, ".yaml", conf2))
	a.NotNil(conf2)
	a.Equal(conf1, conf2)
}

func TestNewContext(t *testing.T) {
	a := assert.New(t)
	initApp(a)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept", "application/json")
	ctx := NewContext(w, r)
	a.NotNil(ctx).
		Equal(ctx.Response, w).
		Equal(ctx.Request, r).
		Equal(ctx.OutputCharsetName, "utf-8").
		Equal(ctx.OutputMimetypeName, "application/json")
}

func TestGrace(t *testing.T) {
	if runtime.GOOS == "windows" { // windows 不支持 os.Process.Signal
		return
	}

	a := assert.New(t)
	exit := make(chan bool, 1)
	initApp(a)

	Grace(300*time.Millisecond, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		Serve()
		exit <- true
	}()
	time.Sleep(300 * time.Microsecond)

	p, err := os.FindProcess(os.Getpid())
	a.NotError(err).NotNil(p)
	a.NotError(p.Signal(syscall.SIGTERM))

	<-exit
}
