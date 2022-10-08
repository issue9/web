// SPDX-License-Identifier: MIT

// Package server 服务管理
package server

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"sync"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/logs/v4"
	"github.com/issue9/mux/v7/group"
	"github.com/issue9/sliceutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/internal/base"
	"github.com/issue9/web/internal/encoding"
	"github.com/issue9/web/internal/serialization"
	"github.com/issue9/web/serializer"
	"github.com/issue9/web/service"
)

// Server web 服务对象
type Server struct {
	name       string
	version    string
	fs         fs.FS
	httpServer *http.Server
	vars       *sync.Map
	cache      cache.Cache
	uptime     time.Time
	routers    *Routers
	problems   *Problems
	services   *service.Server
	base       *base.Base

	closed chan struct{}
	closes []func() error

	beforeMarshals map[string]BeforeMarshalFunc
	afterMarshals  map[string]AfterMarshalFunc
	mimetypes      *serialization.Mimetypes
	encodings      *encoding.Encodings
	files          *serialization.FS
}

// New 返回 *Server 实例
//
// name, version 表示服务的名称和版本号；
// o 指定了初始化 Server 一些非必要参数。
func New(name, version string, o *Options) (*Server, error) {
	const mimetypesCap = 10

	o, err := sanitizeOptions(o)
	if err != nil {
		return nil, err
	}

	loc := base.New(o.Logs, o.Location, o.LanguageTag)
	srv := &Server{
		name:       name,
		version:    version,
		fs:         o.FS,
		httpServer: o.HTTPServer,
		vars:       &sync.Map{},
		cache:      o.Cache,
		uptime:     time.Now(),
		problems:   newProblems(o.ProblemBuilder),
		services:   service.InternalNewServer(loc),
		base:       loc,

		closed: make(chan struct{}, 1),
		closes: make([]func() error, 0, 10),

		beforeMarshals: make(map[string]BeforeMarshalFunc, mimetypesCap),
		afterMarshals:  make(map[string]AfterMarshalFunc, mimetypesCap),
		mimetypes:      serialization.NewMimetypes(mimetypesCap),
		encodings:      encoding.NewEncodings(o.Logs.ERROR()),
		files:          serialization.NewFS(5),
	}
	srv.routers = group.NewOf(srv.call,
		notFound,
		buildNodeHandle(http.StatusMethodNotAllowed),
		buildNodeHandle(http.StatusOK),
		o.RoutersOptions...)
	srv.httpServer.Handler = srv.routers

	return srv, nil
}

// Name 应用的名称
func (srv *Server) Name() string { return srv.name }

// Version 应用的版本
func (srv *Server) Version() string { return srv.version }

func (srv *Server) Open(name string) (fs.File, error) { return srv.fs.Open(name) }

// Vars 操纵共享变量的接口
func (srv *Server) Vars() *sync.Map { return srv.vars }

// Location 指定服务器的时区信息
func (srv *Server) Location() *time.Location { return srv.base.Location }

// Logs 返回关联的 [logs.Logs] 实例
func (srv *Server) Logs() *logs.Logs { return srv.base.Logs }

// Cache 返回缓存的相关接口
func (srv *Server) Cache() cache.Cache { return srv.cache }

// Uptime 当前服务的运行时间
func (srv *Server) Uptime() time.Time { return srv.uptime }

// Mimetypes 编解码控制
func (srv *Server) Mimetypes() serializer.Serializer { return srv.mimetypes }

// Now 返回当前时间
//
// 与 [time.Now] 的区别在于 Now() 基于当前时区
func (srv *Server) Now() time.Time { return time.Now().In(srv.Location()) }

// ParseTime 分析基于当前时区的时间
func (srv *Server) ParseTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, srv.Location())
}

// Serve 启动服务
//
// 会等待 [Server.Close] 执行完之后，此函数才会返回，这一点与 [http.ListenAndServe] 稍有不同。
// 一旦返回，整个 Server 对象将处于不可用状态。
func (srv *Server) Serve() (err error) {
	srv.services.Run()

	// 在 Serve.defer 中关闭服务，而不是 Close。这样可以保证在所有的请求关闭之后执行。
	defer func() {
		srv.services.Stop()

		sliceutil.Reverse(srv.closes)
		for _, f := range srv.closes {
			if err1 := f(); err1 != nil { // 出错不退出，继续其它操作。
				srv.Logs().Error(err1)
			}
		}
	}()

	cfg := srv.httpServer.TLSConfig
	if cfg != nil && (len(cfg.Certificates) > 0 || cfg.GetCertificate != nil) {
		err = srv.httpServer.ListenAndServeTLS("", "")
	} else {
		err = srv.httpServer.ListenAndServe()
	}

	// 由 Server.Close() 主动触发的关闭事件，才需要等待其执行完成，
	// 其它错误直接返回，否则一些内部错误会永远卡在此处无法返回。
	if errors.Is(err, http.ErrServerClosed) {
		<-srv.closed
	}
	return err
}

// Close 触发关闭操作
//
// 需要等待 [Server.Serve] 返回才能证明整个服务被关闭。
func (srv *Server) Close(shutdownTimeout time.Duration) error {
	defer func() {
		srv.closed <- struct{}{}
	}()

	if shutdownTimeout == 0 {
		return srv.httpServer.Close()
	}

	c, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.httpServer.Shutdown(c); !errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return nil
}

// Server 获取关联的 Server 实例
func (ctx *Context) Server() *Server { return ctx.server }

// Files 返回用于序列化文件内容的操作接口
func (srv *Server) Files() serializer.FS { return srv.files }

// LoadLocale 加载本地化的文件
func (srv *Server) LoadLocale(fsys fs.FS, glob string) error {
	if fsys == nil {
		fsys = srv
	}
	return srv.base.LoadLocaleFiles(fsys, glob, srv.files)
}

func (srv *Server) NewPrinter(tag language.Tag) *message.Printer {
	return srv.base.NewPrinter(tag)
}

func (srv *Server) CatalogBuilder() *catalog.Builder { return srv.base.Catalog }

func (srv *Server) LocalePrinter() *message.Printer { return srv.base.Printer }

// LanguageTag 返回默认的语言标签
func (srv *Server) LanguageTag() language.Tag { return srv.base.Tag }

// OnClose 注册关闭服务时需要执行的函数
//
// NOTE: 按注册的相反顺序执行。
func (srv *Server) OnClose(f ...func() error) { srv.closes = append(srv.closes, f...) }

// Services 服务管理
func (srv *Server) Services() *service.Server { return srv.services }
