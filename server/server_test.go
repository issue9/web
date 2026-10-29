// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package server

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v6"
	"github.com/issue9/cache"
	"github.com/issue9/cache/drivers/memory"
	"github.com/issue9/logs/v7"
	"github.com/issue9/mux/v10"
	"golang.org/x/text/language"

	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/json"
	"github.com/issue9/web/mimetype/xml"
	"github.com/issue9/web/selector"
	"github.com/issue9/web/server/registry"
	"github.com/issue9/web/startuptest"
)

var (
	_ web.Startup = &gatewayStartup{}
	_ web.Startup = &serviceStartup{}
)

func buildHandler(code int) web.HandlerFunc {
	return func(ctx *web.Context) web.Responser {
		return web.ResponserFunc(func(ctx *web.Context) { ctx.Render(code, code) })
	}
}

func TestNewHTTP(t *testing.T) {
	a := assert.New(t, false)

	srv, err := NewHTTP("app", "0.1.0", nil)
	a.NotError(err).NotNil(srv).
		False(srv.Uptime().IsZero()).
		NotNil(srv.Cache()).
		Equal(srv.Location(), time.Local)

	d, ok := srv.Cache().(cache.Driver)
	a.True(ok).
		NotNil(d).
		True(srv.Codec().CanCompress())
	srv.Codec().SetCompress(false)
	a.False(srv.Codec().CanCompress())
}

func newOptions(a *assert.Assertion, o *Options) *Options {
	if o == nil {
		o = &Options{Startup: startuptest.NewHTTPServerStartup(a, false), Language: language.English} // 指定不存在的语言
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

func newService(a *assert.Assertion, name string, reg registry.Registry, c cache.Driver) (*web.Server, startuptest.Startup) {
	startup := startuptest.NewHTTPServerStartup(a, false)
	srv, err := NewService(name, "0.1.0", newOptions(a, &Options{
		Cache:    c,
		Startup:  startup,
		Registry: reg,
		Peer:     selector.NewPeer(startup.URL()),
	}))
	a.NotError(err).NotNil(srv)

	return srv, startup
}

func TestNewService(t *testing.T) {
	a := assert.New(t, false)

	// Registry 和 Peer 是空的
	srv, err := NewService("app", "0.1.0", newOptions(a, nil))
	a.Error(err).Nil(srv)

	c := memory.New()
	reg := registry.NewCache(web.NewCache("reg:", c), registry.NewRandomStrategy(), time.Second)

	srv, startup := newService(a, "app", reg, c)
	a.Equal(srv.ID(), "app").Equal(srv.Version(), "0.1.0")

	srv.Routers().New("default", nil).Get("/mux/test", buildHandler(202))

	go srv.Serve()
	defer srv.Close(0)

	startup.Get("/mux/test").Do().Status(202)
}

func TestNewGateway(t *testing.T) {
	a := assert.New(t, false)
	c := memory.New() // 默认的缓存系统用的是内存类型的，保证引用同一个。

	reg := registry.NewCache(web.NewCache("reg:", c), registry.NewRandomStrategy(), time.Second)

	// s1
	s1, _ := newService(a, "s1", reg, c)
	go s1.Serve()
	defer s1.Close(0)
	s1.Routers().New("default", nil).Get("/mux/test", buildHandler(201))

	// s2
	s2, _ := newService(a, "s2", reg, c)
	go s2.Serve()
	defer s2.Close(0)
	s2.Routers().New("default", nil).Get("/mux/test", buildHandler(202))

	// Registry 和 Mapper 是空的
	g, err := NewGateway("app", "0.1.0", newOptions(a, nil))
	a.Equal(err.(*web.FieldError).Field, "o.Mapper").Nil(g)

	sg := startuptest.NewServerStartup(a)
	g, err = NewGateway("app", "0.1.0", newOptions(a, &Options{
		Cache:    c,
		Startup:  sg,
		Registry: reg,
		Mapper: map[string]mux.Matcher{
			"s1": mux.NewPathVersion("", "/s1"),
			"s2": mux.NewPathVersion("", "/s2"),
		},
	}))
	a.NotError(err).NotNil(g)
	a.Equal(g.ID(), "app").Equal(g.Version(), "0.1.0")
	g.Routers().New("default", nil).Get("/mux/test", buildHandler(203))

	go g.Serve()
	defer g.Close(0)

	sg.Get("/s1/mux/test").Do().Status(201)
	sg.Get("/s2/mux/test").Do().Status(202)
	sg.Get("/s3/mux/test").Do().Status(http.StatusNotFound)
	sg.Get("/mux/test").Do().Status(203)
}
