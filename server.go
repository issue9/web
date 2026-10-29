// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"context"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/config"
	"github.com/issue9/mux/v10"
	"golang.org/x/text/language"

	"github.com/issue9/web/selector"
)

type (
	// OnExitContextFunc 表示在即将退出 [Context] 时执行的函数类型
	//
	// 其中 ctx 即为当前实例，status 则表示实际输出的状态码。
	OnExitContextFunc = func(ctx *Context, status int)

	// Plugin 附加于 [Server] 的插件
	Plugin interface {
		Plugin(*Server)
	}

	PluginFunc func(*Server)

	// Server 服务接口
	//
	// 创建该接口对象的方法在 [server] 包之下，提供了不同功能的服务对象。
	Server struct {
		startup Startup
		state   State
		id      string
		version string

		locale   *Locale
		location *time.Location
		uptime   time.Time

		done    chan struct{}
		doneErr error

		config       *config.Config
		requestIDKey string
		codec        *Codec
		services     *Services
		problems     *Problems
		routers      *Routers
		vars         *sync.Map
		idgen        func() string
		logs         *Logs
		closes       []func() error
		cache        cache.Driver
		exitContexts []OnExitContextFunc
	}
)

// NewServer 创建 [Server]
//
// requestIDKey 表示客户端提交的 X-Request-ID 报头名；
func NewServer(
	id, ver string,
	conf *config.Config,
	startup Startup,
	loc *time.Location,
	lang language.Tag,
	logs *Logs,
	idgen func() string,
	cache cache.Driver,
	codec *Codec,
	requestIDKey string,
	problemTypePrefix string,
	o ...RouterOption,
) *Server {
	srv := &Server{
		startup: startup,
		id:      id,
		version: ver,

		locale:   newLocale(lang, conf.Serializer()),
		location: loc,
		uptime:   time.Now().In(loc),

		done: make(chan struct{}),

		config:       conf,
		requestIDKey: requestIDKey,
		codec:        codec,
		problems:     newProblems(problemTypePrefix),
		vars:         &sync.Map{},
		idgen:        idgen,
		logs:         logs,
		closes:       make([]func() error, 0, 10),
		cache:        cache,
		exitContexts: make([]OnExitContextFunc, 0, 10),
	}
	srv.initServices()
	srv.Logs().SetLocale(srv.Locale().Printer())

	o = append(o, withRecovery(srv))
	srv.routers = &Routers{
		g: mux.NewGroup(srv.call,
			notFound,
			trace,
			buildNodeHandle(http.StatusMethodNotAllowed),
			buildNodeHandle(http.StatusOK),
			o...),
	}
	srv.startup.Init(srv.routers.g)

	return srv
}

// NewClient 基于当前对象的相关字段创建 [Client] 对象
//
// 功能与 [NewClient] 相同，缺少的参数直接采用 [Server] 关联的字段。
func (s *Server) NewClient(c *http.Client, sel selector.Selector, m string, marshal func(any) ([]byte, error)) *Client {
	return NewClient(c, s.Codec(), sel, m, marshal, s.requestIDKey, s.UniqueID)
}

// Config 当前项目配置文件的管理
func (s *Server) Config() *config.Config { return s.config }

// Locale 提供本地化相关功能
func (s *Server) Locale() *Locale { return s.locale }

// ID 应用的 ID
func (s *Server) ID() string { return s.id }

// Version 应用的版本
func (s *Server) Version() string { return s.version }

// Vars 操纵共享变量的接口
func (s *Server) Vars() *sync.Map { return s.vars }

// Location 服务器的时区信息
func (s *Server) Location() *time.Location { return s.location }

// Now 返回当前时间
//
// 与 [time.Now] 的区别在于 Now 的时间基于 [Server.Location]
func (s *Server) Now() time.Time { return time.Now().In(s.Location()) }

// Uptime 当前服务的运行时间
func (s *Server) Uptime() time.Time { return s.uptime }

// Cache 返回缓存的相关接口
//
// 如果要获得缓存的底层驱动接口，可以将类型转换为 [cache.Driver]，
// 该类型提供了 [cache.Driver.Driver] 方法可以获得相应的对象。
func (s *Server) Cache() cache.Cleanable { return s.cache }

func (s *Server) Codec() *Codec { return s.codec }

// ParseTime 分析基于当前时区的时间
func (s *Server) ParseTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, s.Location())
}

// UniqueID 生成唯一性的 ID
func (s *Server) UniqueID() string { return s.idgen() }

// OnClose 注册关闭服务时需要执行的函数
//
// 如果参数 f 支持 [context.Context]，可以直接使用 [context.WithCancel] 直接从 Server 订阅关闭事件。
//
// NOTE: 按注册的相反顺序执行。
func (s *Server) OnClose(f ...func() error) { s.closes = append(s.closes, f...) }

// OnExitContext 注册在即将退出 [Context] 时需要执行的函数
func (s *Server) OnExitContext(f ...OnExitContextFunc) {
	s.exitContexts = append(s.exitContexts, f...)
}

// Logs 提供日志接口
func (s *Server) Logs() *Logs { return s.logs }

func (s *Server) Deadline() (time.Time, bool) { return time.Time{}, false }

func (s *Server) Value(key any) any {
	if val, found := s.Vars().Load(key); found {
		return val
	}
	return nil
}

func (s *Server) Done() <-chan struct{} { return s.done }

func (s *Server) Err() error { return s.doneErr }

func (f PluginFunc) Plugin(s *Server) { f(s) }

// Use 添加插件
func (s *Server) Use(p ...Plugin) {
	for _, pp := range p {
		pp.Plugin(s)
	}
}

// State 获取当前的状态
func (s *Server) State() State { return s.state }

// Serve 开始 HTTP 服务
//
// 这是个阻塞方法，会等待 [Server.Close] 执行完之后才返回。
// 始终返回非空的错误对象，如果是由 [Server.Close] 关闭的，将返回 [http.ErrServerClosed]。
func (s *Server) Serve() (err error) {
	if s.State() == Running {
		panic("当前已经处于运行状态")
	}
	s.state = Running

	err = s.startup.Serve(s)
	<-s.Done()
	return err
}

// Close 关闭服务
//
// 在执行完之后会通知 Serve 返回。
//
// 调用此方法表示 [Server] 的生命周期结束，对象将处于不可用状态。
func (s *Server) Close(shutdownTimeout time.Duration) {
	if s.State() != Running {
		return
	}
	s.state = Stopped // 调用 Close 即设置状态

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)

	defer func() {
		s.close()
		cancel()
	}()

	if err := s.startup.Close(ctx); err != nil {
		s.Logs().ERROR().Error(err)
	}
}

func (s *Server) close() {
	for _, f := range slices.Backward(s.closes) {
		if err := f(); err != nil { // 出错不退出，继续其它操作。
			s.Logs().ERROR().Error(err)
		}
	}

	s.doneErr = http.ErrServerClosed // 在 close 之前调用，以保证在 close 之后，doneErr 始终是正确的。
	close(s.done)
}
