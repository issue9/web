// SPDX-License-Identifier: MIT

//go:generate go run ./make_problems.go

// Package server 服务端实现
package server

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/issue9/config"
	"github.com/issue9/mux/v7/group"

	"github.com/issue9/web"
	"github.com/issue9/web/cache"
	"github.com/issue9/web/internal/locale"
)

type httpServer struct {
	name        string
	version     string
	httpServer  *http.Server
	vars        *sync.Map
	cache       cache.Driver
	uptime      time.Time
	routers     *group.GroupOf[web.HandlerFunc]
	idGenerator IDGenerator
	state       web.State
	services    *services
	ctxBuilder  *web.ContextBuilder
	location    *time.Location
	logs        web.Logs

	closed chan struct{}
	closes []func() error

	problems        *problems
	codec           *web.Codec
	config          *config.Config
	locale          *locale.Locale
	disableCompress bool
}

// New 新建 http 服务
//
// name, version 表示服务的名称和版本号；
// o 指定了一些带有默认值的参数；
func New(name, version string, o *Options) (web.Server, error) {
	o, err := sanitizeOptions(o)
	if err != nil {
		err.Path = "Options"
		return nil, err
	}

	srv := &httpServer{
		name:        name,
		version:     version,
		httpServer:  o.HTTPServer,
		vars:        &sync.Map{},
		cache:       o.Cache,
		uptime:      time.Now(),
		idGenerator: o.IDGenerator,
		state:       web.Stopped,

		location: o.Location,
		logs:     o.logs,

		closed: make(chan struct{}, 1),
		closes: make([]func() error, 0, 10),

		problems: o.problems,
		codec:    o.codec,
		locale:   o.locale,
		config:   o.Config,
	}

	srv.ctxBuilder = web.NewContextBuilder(srv, o.codec, o.RequestIDKey)

	srv.routers = group.NewOf(srv.call,
		notFound,
		buildNodeHandle(http.StatusMethodNotAllowed),
		buildNodeHandle(http.StatusOK),
		o.RoutersOptions...)
	srv.httpServer.Handler = srv.routers
	srv.OnClose(srv.cache.Close)
	srv.initServices()

	for _, f := range o.Init { // NOTE: 需要保证在最后
		f(srv)
	}
	return srv, nil
}

func (srv *httpServer) Name() string { return srv.name }

func (srv *httpServer) Version() string { return srv.version }

func (srv *httpServer) State() web.State { return srv.state }

func (srv *httpServer) Vars() *sync.Map { return srv.vars }

func (srv *httpServer) Location() *time.Location { return srv.location }

func (srv *httpServer) Cache() cache.Cleanable { return srv.cache }

func (srv *httpServer) Uptime() time.Time { return srv.uptime }

func (srv *httpServer) UniqueID() string { return srv.idGenerator() }

func (srv *httpServer) Now() time.Time { return time.Now().In(srv.Location()) }

func (srv *httpServer) ParseTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, srv.Location())
}

func (srv *httpServer) Serve() (err error) {
	if srv.State() == web.Running {
		panic("当前已经处于运行状态")
	}
	srv.state = web.Running

	cfg := srv.httpServer.TLSConfig
	if cfg != nil && (len(cfg.Certificates) > 0 || cfg.GetCertificate != nil) {
		err = srv.httpServer.ListenAndServeTLS("", "")
	} else {
		err = srv.httpServer.ListenAndServe()
	}

	if errors.Is(err, http.ErrServerClosed) {
		// 由 [Server.Close] 主动触发的关闭事件，才需要等待其执行完成。
		// 其它错误直接返回，否则一些内部错误会永远卡在此处无法返回。
		<-srv.closed
	}
	return err
}

func (srv *httpServer) Close(shutdownTimeout time.Duration) {
	if srv.State() != web.Running {
		return
	}

	defer func() {
		slices.Reverse(srv.closes)
		for _, f := range srv.closes { // 仅在用户主动要关闭时，才关闭服务。
			if err1 := f(); err1 != nil { // 出错不退出，继续其它操作。
				srv.Logs().ERROR().Error(err1)
			}
		}

		srv.state = web.Stopped
		srv.closed <- struct{}{} // NOTE: 保证最后执行
	}()

	if shutdownTimeout == 0 {
		if err := srv.httpServer.Close(); err != nil {
			srv.Logs().ERROR().Error(err)
		}
		return
	}

	c, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.httpServer.Shutdown(c); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		srv.Logs().ERROR().Error(err)
	}
}

func (srv *httpServer) OnClose(f ...func() error) { srv.closes = append(srv.closes, f...) }

func (srv *httpServer) Logs() web.Logs { return srv.logs }

func (srv *httpServer) Config() *config.Config { return srv.config }

func (srv *httpServer) Locale() web.Locale { return srv.locale }

func (srv *httpServer) NewClient(client *http.Client, selector web.Selector, marshalName string, marshal func(any) ([]byte, error)) *web.Client {
	return web.NewClient(client, srv.codec, selector, marshalName, marshal)
}

func (srv *httpServer) CanCompress() bool { return !srv.disableCompress }

func (srv *httpServer) SetCompress(enable bool) { srv.disableCompress = !enable }
