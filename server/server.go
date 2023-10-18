// SPDX-License-Identifier: MIT

//go:generate go run ./make_problems.go

// Package server 服务端实现
package server

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"sync"
	"time"

	"github.com/issue9/config"
	"github.com/issue9/mux/v7/group"
	"github.com/issue9/sliceutil"
	"golang.org/x/text/encoding"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web"
	"github.com/issue9/web/cache"
	"github.com/issue9/web/internal/compress"
	"github.com/issue9/web/internal/locale"
	"github.com/issue9/web/internal/mimetypes"
	"github.com/issue9/web/internal/problems"
)

type httpServer struct {
	name         string
	version      string
	httpServer   *http.Server
	vars         *sync.Map
	cache        cache.Driver
	uptime       time.Time
	routers      *group.GroupOf[web.HandlerFunc]
	idGenerator  IDGenerator
	requestIDKey string
	state        web.State
	services     *services

	location *time.Location
	catalog  *catalog.Builder
	tag      language.Tag
	printer  *message.Printer
	logs     web.Logs

	closed chan struct{}
	closes []func() error

	problems   *problems.Problems
	mimetypes  *mimetypes.Mimetypes
	compresses *compress.Compresses
	config     *config.Config
}

// New 新建 http 服务
//
// name, version 表示服务的名称和版本号；
// o 指定了初始化 [web.Server] 一些带有默认值的参数；
func New(name, version string, o *Options) (web.Server, error) {
	o, err := sanitizeOptions(o)
	if err != nil {
		err.Path = "Options"
		return nil, err
	}

	srv := &httpServer{
		name:         name,
		version:      version,
		httpServer:   o.HTTPServer,
		vars:         &sync.Map{},
		cache:        o.Cache,
		uptime:       time.Now(),
		idGenerator:  o.IDGenerator,
		requestIDKey: o.RequestIDKey,
		state:        web.Stopped,

		location: o.Location,
		catalog:  o.Catalog,
		tag:      o.Language,
		printer:  o.printer,
		logs:     o.logs,

		closed: make(chan struct{}, 1),
		closes: make([]func() error, 0, 10),

		problems:   o.problems,
		mimetypes:  o.mimetypes,
		compresses: o.compresses,
		config:     o.Config,
	}

	initProblems(srv.problems)

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
		// 由 Server.Close() 主动触发的关闭事件，才需要等待其执行完成。
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
		sliceutil.Reverse(srv.closes)  // TODO: go1.21 改为标准库
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

func (srv *httpServer) NewLocalePrinter(tag language.Tag) *message.Printer {
	return newPrinter(tag, srv.Catalog())
}

func (srv *httpServer) Catalog() *catalog.Builder { return srv.catalog }

func (srv *httpServer) LocalePrinter() *message.Printer { return srv.printer }

func (srv *httpServer) Language() language.Tag { return srv.tag }

func (srv *httpServer) LoadLocale(glob string, fsys ...fs.FS) error {
	return locale.Load(srv.Config().Serializer(), srv.Catalog(), glob, fsys...)
}

func (srv *httpServer) DisableCompress(disable bool) { srv.compresses.SetDisable(disable) }

func (srv *httpServer) CompressIsDisable() bool { return srv.compresses.IsDisable() }

func (srv *httpServer) ContentType(h string) (web.UnmarshalFunc, encoding.Encoding, error) {
	return srv.mimetypes.ContentType(h)
}

func (srv *httpServer) Accept(h string) *web.Mimetype {
	m := srv.mimetypes.Accept(h)
	if m == nil {
		return nil
	}
	return &web.Mimetype{
		Name:           m.Name,
		Problem:        m.Problem,
		MarshalBuilder: m.MarshalBuilder,
		Unmarshal:      m.Unmarshal,
	}
}

func (srv *httpServer) ContentEncoding(name string, r io.Reader) (io.ReadCloser, error) {
	return srv.compresses.ContentEncoding(name, r)
}

func (srv *httpServer) AcceptEncoding(contentType, h string, l web.Logger) (w web.Compressor, name string, notAcceptable bool) {
	n, not := srv.compresses.AcceptEncoding(contentType, h, l)
	if not {
		return nil, "", true
	}
	if n == nil {
		return nil, "", false
	}
	return n.Compress(), n.Name(), false
}
