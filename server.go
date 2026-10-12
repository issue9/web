// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/config"
	"github.com/issue9/mux/v10"
	"github.com/issue9/mux/v10/types"
	"golang.org/x/text/language"

	"github.com/issue9/web/selector"
)

type (
	// Server 服务接口
	//
	// 创建该接口对象的方法在 [server] 包之下，提供了不同功能的服务对象。
	Server interface {
		context.Context

		// ID 应用的 ID
		ID() string

		// Version 应用的版本
		Version() string

		// State 获取当前的状态
		State() State

		// Vars 操纵共享变量的接口
		Vars() *sync.Map

		// Location 服务器的时区信息
		Location() *time.Location

		// Cache 返回缓存的相关接口
		//
		// 如果要获得缓存的底层驱动接口，可以将类型转换为 [cache.Driver]，
		// 该类型提供了 [cache.Driver.Driver] 方法可以获得相应的对象。
		Cache() cache.Cleanable

		Codec() *Codec

		// Uptime 当前服务的运行时间
		Uptime() time.Time

		// UniqueID 生成唯一性的 ID
		UniqueID() string

		// Now 返回当前时间
		//
		// 与 [time.Now] 的区别在于 Now 的时间基于 [Server.Location]
		Now() time.Time

		// ParseTime 分析基于当前时区的时间
		ParseTime(layout, value string) (time.Time, error)

		// Serve 开始 HTTP 服务
		//
		// 这是个阻塞方法，会等待 [Server.Close] 执行完之后才返回。
		// 始终返回非空的错误对象，如果是由 [Server.Close] 关闭的，将返回 [http.ErrServerClosed]。
		Serve() error

		// Close 关闭服务
		//
		// 在执行完之后会通知 Serve 返回。
		//
		// 调用此方法表示 [Server] 的生命周期结束，对象将处于不可用状态。
		Close(shutdownTimeout time.Duration)

		// OnClose 注册关闭服务时需要执行的函数
		//
		// NOTE: 按注册的相反顺序执行。
		OnClose(...func() error)

		// OnExitContext 注册在即将退出 [Context] 时需要执行的函数
		OnExitContext(...OnExitContextFunc)

		// Config 当前项目配置文件的管理
		Config() *config.Config

		// Logs 提供日志接口
		Logs() *Logs

		// NewContext 从标准库的参数初始化 [Context] 对象
		//
		// 这适合从标准库的请求中创建 [Context] 对象。
		// [types.Route] 类型的参数需要用户通过 [types.NewRoute] 自行创建。
		//
		// NOTE: 由此方法创建的对象在整个会话结束后会被回收。
		NewContext(http.ResponseWriter, *http.Request, *types.Route) *Context

		// NewClient 基于当前对象的相关字段创建 [Client] 对象
		//
		// 功能与 [NewClient] 相同，缺少的参数直接采用 [Server] 关联的字段。
		NewClient(client *http.Client, s selector.Selector, marshalName string, marshal func(any) ([]byte, error)) *Client

		// Routers 路由管理
		Routers() *Routers

		// Problems 提供 [Problem] 管理
		Problems() *Problems

		// Services 服务管理接口
		Services() *Services

		// Locale 提供本地化相关功能
		Locale() *Locale

		// Use 添加插件
		Use(...Plugin)
	}

	// OnExitContextFunc 表示在即将退出 [Context] 时执行的函数类型
	//
	// 其中 ctx 即为当前实例，status 则表示实际输出的状态码。
	OnExitContextFunc = func(ctx *Context, status int)

	// Plugin 附加于 [Server] 的插件
	Plugin interface {
		Plugin(Server)
	}

	PluginFunc func(Server)

	OnRender func(status int, body any) (int, any)

	serverImplement struct {
		hs    *http.Server
		state State

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
		onRender     OnRender
		exitContexts []OnExitContextFunc
	}
)

// NewServer 创建 [Server]
//
// requestIDKey 表示客户端提交的 X-Request-ID 报头名；
// onRender 在每个对象的渲染之前可以对内容进行的修改；
func NewServer(
	id, ver string,
	conf *config.Config,
	hs *http.Server,
	loc *time.Location,
	lang language.Tag,
	logs *Logs,
	idgen func() string,
	cache cache.Driver,
	codec *Codec,
	requestIDKey string,
	problems *Problems,
	onRender OnRender,
	o ...RouterOption,
) Server {
	impl := &serverImplement{
		hs:    hs,
		state: Stopped,

		id:      id,
		version: ver,

		locale:   newLocale(lang, conf.Serializer()),
		location: loc,
		uptime:   time.Now().In(loc),

		done: make(chan struct{}),

		config:       conf,
		requestIDKey: requestIDKey,
		codec:        codec,
		problems:     problems,
		vars:         &sync.Map{},
		idgen:        idgen,
		logs:         logs,
		closes:       make([]func() error, 0, 10),
		cache:        cache,
		onRender:     onRender,
		exitContexts: make([]OnExitContextFunc, 0, 10),
	}
	impl.initServices()
	impl.Logs().SetLocale(impl.Locale().Printer())

	o = append(o, withRecovery(impl))
	impl.routers = &Routers{
		g: mux.NewGroup(impl.call,
			notFound,
			trace,
			buildNodeHandle(http.StatusMethodNotAllowed),
			buildNodeHandle(http.StatusOK),
			o...),
	}
	impl.hs.Handler = impl.routers.g

	return impl
}

func (s *serverImplement) NewClient(c *http.Client, sel selector.Selector, m string, marshal func(any) ([]byte, error)) *Client {
	return NewClient(c, s.codec, sel, m, marshal, s.requestIDKey, s.UniqueID)
}

func (s *serverImplement) Config() *config.Config { return s.config }

func (s *serverImplement) Locale() *Locale { return s.locale }

func (s *serverImplement) ID() string { return s.id }

func (s *serverImplement) Version() string { return s.version }

func (s *serverImplement) Vars() *sync.Map { return s.vars }

func (s *serverImplement) Location() *time.Location { return s.location }

func (s *serverImplement) Now() time.Time { return time.Now().In(s.Location()) }

func (s *serverImplement) Uptime() time.Time { return s.uptime }

func (s *serverImplement) Cache() cache.Cleanable { return s.cache }

func (s *serverImplement) Codec() *Codec { return s.codec }

func (s *serverImplement) ParseTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, s.Location())
}

func (s *serverImplement) UniqueID() string { return s.idgen() }

func (s *serverImplement) OnClose(f ...func() error) { s.closes = append(s.closes, f...) }

func (s *serverImplement) OnExitContext(f ...OnExitContextFunc) {
	s.exitContexts = append(s.exitContexts, f...)
}

func (s *serverImplement) Logs() *Logs { return s.logs }

func (s *serverImplement) close() {
	for _, f := range slices.Backward(s.closes) {
		if err := f(); err != nil { // 出错不退出，继续其它操作。
			s.Logs().ERROR().Error(err)
		}
	}

	s.doneErr = http.ErrServerClosed // 在 close 之前调用，以保证在 close 之后，doneErr 始终是正确的。
	close(s.done)
}

func (s *serverImplement) Deadline() (time.Time, bool) { return time.Time{}, false }

func (s *serverImplement) Value(key any) any {
	if val, found := s.Vars().Load(key); found {
		return val
	}
	return nil
}

func (s *serverImplement) Done() <-chan struct{} { return s.done }

func (s *serverImplement) Err() error { return s.doneErr }

func (f PluginFunc) Plugin(s Server) { f(s) }

func (s *serverImplement) Use(p ...Plugin) {
	for _, pp := range p {
		pp.Plugin(s)
	}
}

func (s *serverImplement) State() State { return s.state }

func (s *serverImplement) Serve() (err error) {
	if s.State() == Running {
		panic("当前已经处于运行状态")
	}
	s.state = Running

	if c := s.hs.TLSConfig; c != nil && (len(c.Certificates) > 0 || c.GetCertificate != nil) {
		err = s.hs.ListenAndServeTLS("", "")
	} else {
		err = s.hs.ListenAndServe()
	}

	<-s.Done()
	return err
}

func (s *serverImplement) Close(shutdownTimeout time.Duration) {
	if s.State() != Running {
		return
	}
	s.state = Stopped // 调用 Close 即设置状态

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)

	defer func() {
		s.close()
		cancel()
	}()

	if err := s.hs.Shutdown(ctx); err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		s.Logs().ERROR().Error(err)
	}
}
