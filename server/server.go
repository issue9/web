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

	"github.com/issue9/mux/v7/group"
	"github.com/issue9/sliceutil"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/cache"
	"github.com/issue9/web/internal/encoding"
	"github.com/issue9/web/internal/files"
	"github.com/issue9/web/internal/mimetypes"
	"github.com/issue9/web/internal/problems"
	"github.com/issue9/web/logs"
)

// Server web 服务对象
type Server struct {
	name            string
	version         string
	fs              fs.FS
	httpServer      *http.Server
	vars            *sync.Map
	cache           cache.Driver
	uptime          time.Time
	routers         *group.GroupOf[HandlerFunc]
	uniqueGenerator func() string
	requestIDKey    string
	state           State
	services        *Services

	location *time.Location
	catalog  *catalog.Builder
	tag      language.Tag
	printer  *message.Printer
	logs     *logs.Logs

	closed chan struct{}
	closes []func() error

	problems  *problems.Problems[Problem]
	mimetypes *mimetypes.Mimetypes[MarshalFunc, UnmarshalFunc]
	encodings *encoding.Encodings
	files     *Files
}

type Files = files.Files

// New 新建 web 服务
//
// name, version 表示服务的名称和版本号；
// o 指定了初始化 [Server] 一些带有默认值的参数。
func New(name, version string, o *Options) (*Server, error) {
	o, err := sanitizeOptions(o)
	if err != nil {
		err.Path = "server.Options"
		return nil, err
	}

	srv := &Server{
		name:            name,
		version:         version,
		fs:              o.FS,
		httpServer:      o.HTTPServer,
		vars:            &sync.Map{},
		cache:           o.Cache,
		uptime:          time.Now(),
		uniqueGenerator: o.UniqueGenerator.String,
		requestIDKey:    o.RequestIDKey,
		state:           Stopped,

		location: o.Location,
		catalog:  o.Locale.Catalog,
		tag:      o.Locale.Language,
		printer:  o.Locale.printer,
		logs:     o.logs,

		closed: make(chan struct{}, 1),
		closes: make([]func() error, 0, 10),

		problems:  o.problems,
		mimetypes: o.mimetypes,
		encodings: encoding.NewEncodings(o.logs.ERROR()),
	}

	for _, e := range o.Encodings {
		srv.encodings.Add(e.Name, e.Builder, e.ContentTypes...)
	}
	srv.files = files.New(srv)
	srv.routers = group.NewOf(srv.call,
		notFound,
		buildNodeHandle(http.StatusMethodNotAllowed),
		buildNodeHandle(http.StatusOK),
		o.RoutersOptions...)
	srv.httpServer.Handler = srv.routers
	srv.httpServer.ErrorLog = srv.Logs().ERROR().StdLogger()
	srv.OnClose(srv.cache.Close)
	srv.initServices(o.UniqueGenerator) // 依赖 logs 和 Location 已经初始化

	return srv, nil
}

// Name 应用的名称
func (srv *Server) Name() string { return srv.name }

// Version 应用的版本
func (srv *Server) Version() string { return srv.version }

// State 获取当前的状态
//
// 返回值可能会是以下值：
//   - [Running] 表示已经运行 [Server.Serve]；
//   - [Stopped] 默认状态，或由 [Server.Close] 正常结束；
func (srv *Server) State() State { return srv.state }

func (srv *Server) Open(name string) (fs.File, error) { return srv.fs.Open(name) }

// Vars 操纵共享变量的接口
func (srv *Server) Vars() *sync.Map { return srv.vars }

// Location 指定服务器的时区信息
func (srv *Server) Location() *time.Location { return srv.location }

// Cache 返回缓存的相关接口
func (srv *Server) Cache() cache.CleanableCache { return srv.cache }

// Uptime 当前服务的运行时间
func (srv *Server) Uptime() time.Time { return srv.uptime }

// UniqueID 生成唯一性的 ID
func (srv *Server) UniqueID() string { return srv.uniqueGenerator() }

// Now 返回当前时间
//
// 与 [time.Now] 的区别在于 Now() 基于当前时区
func (srv *Server) Now() time.Time { return time.Now().In(srv.Location()) }

// ParseTime 分析基于当前时区的时间
func (srv *Server) ParseTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, srv.Location())
}

// Serve 开始 HTTP 服务
//
// 这是个阻塞方法，会等待 [Server.Close] 执行完之后才返回。
func (srv *Server) Serve() (err error) {
	if srv.State() == Running {
		panic("当前已经处于运行状态")
	}

	srv.state = Running

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

// Close 关闭服务
//
// 无论是否出错，该操作最终都会导致 [Server.Serve] 的退出。
// 调用此方法表示 [Server] 的生命周期结束，对象将处于不可用状态。
func (srv *Server) Close(shutdownTimeout time.Duration) {
	if srv.State() != Running {
		return
	}

	defer func() {
		sliceutil.Reverse(srv.closes)
		for _, f := range srv.closes { // 仅在用户主动要关闭时，才关闭服务。
			if err1 := f(); err1 != nil { // 出错不退出，继续其它操作。
				srv.Logs().ERROR().Error(err1)
			}
		}

		srv.state = Stopped
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
	if err := srv.httpServer.Shutdown(c); !errors.Is(err, context.DeadlineExceeded) {
		srv.Logs().ERROR().Error(err)
		return
	}
}

// Server 获取关联的 Server 实例
func (ctx *Context) Server() *Server { return ctx.server }

func (srv *Server) NewPrinter(tag language.Tag) *message.Printer {
	return newPrinter(tag, srv.CatalogBuilder())
}

func (srv *Server) CatalogBuilder() *catalog.Builder { return srv.catalog }

func (srv *Server) LocalePrinter() *message.Printer { return srv.printer }

// Language 返回默认的语言标签
func (srv *Server) Language() language.Tag { return srv.tag }

// OnClose 注册关闭服务时需要执行的函数
//
// NOTE: 按注册的相反顺序执行。
func (srv *Server) OnClose(f ...func() error) { srv.closes = append(srv.closes, f...) }

// LoadLocales 加载本地化的内容
func (srv *Server) LoadLocales(fsys fs.FS, glob string) error {
	return files.LoadLocales(srv.Files(), srv.CatalogBuilder(), fsys, glob)
}

// Files 配置文件的相关操作
func (srv *Server) Files() *Files { return srv.files }
