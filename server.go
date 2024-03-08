// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"io/fs"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/config"
	"github.com/issue9/mux/v7/group"
	"github.com/issue9/mux/v7/types"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/internal/locale"
	"github.com/issue9/web/selector"
)

type (
	// Server 服务接口
	Server interface {
		// Name 应用的名称
		Name() string

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

		// Close 触发关闭服务事件
		//
		// 只是触发事件，需要等到 [Server.Serve] 返回才表示服务真正结束。
		// 调用此方法表示 [Server] 的生命周期结束，对象将处于不可用状态。
		Close(shutdownTimeout time.Duration)

		// OnClose 注册关闭服务时需要执行的函数
		//
		// NOTE: 按注册的相反顺序执行。
		OnClose(...func() error)

		// OnContext 注册在请求进来之后 [Context] 创建之前可执行的函数
		OnContext(...OnContextFunc)

		// OnExitContext 注册在即将退出 [Context] 时需要执行的函数
		OnExitContext(...OnExitContextFunc)

		// Config 当前项目配置文件的管理
		Config() *config.Config

		// Logs 提供日志接口
		Logs() *Logs

		// NewContext 从标准库的参数初始化 [Context] 对象
		//
		// 这适合从标准库的请求中创建 [Context] 对象。
		// [types.Route] 类型的参数需要用户通过 [types.NewContext] 自行创建。
		//
		// NOTE: 由此方法创建的对象在整个会话结束后会被回收。
		NewContext(http.ResponseWriter, *http.Request, types.Route) *Context

		// NewClient 基于当前对象的相关字段创建 [Client] 对象
		//
		// 功能与 [NewClient] 相同，缺少的参数直接采用 [Server] 关联的字段。
		NewClient(client *http.Client, s selector.Selector, marshalName string, marshal func(any) ([]byte, error)) *Client

		// Routers 路由管理
		Routers() *Routers

		// SetCompress 设置压缩功能
		//
		// 在服务器性能吃紧的情况下可以采用此方法禁用压缩。
		//
		// NOTE: 仅对输出内容启作用，读取内容始终是按照提交的 Content-Encoding 指定算法进行解析。
		SetCompress(enable bool)

		// CanCompress 当前是否拥有压缩功能
		CanCompress() bool

		// Problems 提供 [Problem] 管理
		Problems() *Problems

		// Services 服务管理接口
		Services() *Services

		// Locale 提供本地化相关功能
		Locale() Locale

		// Use 添加插件
		Use(...Plugin)
	}

	// OnContextFunc 表示在创建 [Context] 时执行的函数类型
	//
	// 此时已经获得 [http.ResponseWriter] 和 [http.Request]，但是还没有创建 [Context]，
	// 用户可以根据自己的需求调整 [http.ResponseWriter] 和 [http.Request]。
	OnContextFunc = func(http.ResponseWriter, *http.Request, types.Route) (http.ResponseWriter, *http.Request)

	// OnExitContextFunc 表示在即将退出 [Context] 时执行的函数类型
	//
	// 其中 ctx 即为当前实例，status 则表示实际输出的状态码。
	OnExitContextFunc = func(ctx *Context, status int)

	// Locale 提供与本地化相关的功能
	Locale interface {
		catalog.Catalog

		// ID 返回默认的语言标签
		ID() language.Tag

		// LoadMessages 从 fsys 中加载符合 glob 的本地化文件
		//
		// 根据 [Server.Config] 处理文件格式，如果文件格式不被 [Server.Config] 支持，将无法加载。
		LoadMessages(glob string, fsys ...fs.FS) error

		// Printer 最符合 [Locale.ID] 的 [message.Printer] 对象
		Printer() *message.Printer

		// Sprintf 等同于 [Locale.Printer.Sprintf]
		Sprintf(string, ...any) string

		// NewPrinter 声明最符合 tag 的 [message.Printer] 对象
		//
		// NOTE: 每当给 [Locale.SetString]、[Locale.SetMacro] 和 [Locale.Set] 传递新的 [language.Tag]
		// 值时，可能造成 NewPrinter 相同的入参而返回不同的返回对象的情况。
		NewPrinter(tag language.Tag) *message.Printer

		// SetString 添加新的翻译项
		//
		// 功能同 [catalog.Builder.SetString]
		SetString(tag language.Tag, key, msg string) error

		// SetMacro 添加新的翻译项
		//
		// 功能同 [catalog.Builder.SetMacro]
		SetMacro(tag language.Tag, name string, msg ...catalog.Message) error

		// Set 添加新的翻译项
		//
		// 功能同 [catalog.Builder.Set]
		Set(tag language.Tag, key string, msg ...catalog.Message) error
	}

	// Plugin 附加于 [Server] 的插件
	Plugin interface {
		Plugin(Server)
	}

	PluginFunc func(Server)

	// InternalServer 这是一个内部使用的类型，提供了大部分 [Server] 的实现。
	InternalServer struct {
		server Server

		name    string
		version string

		locale   *locale.Locale
		location *time.Location
		uptime   time.Time

		requestIDKey    string
		codec           *Codec
		services        *Services
		problems        *Problems
		routers         *Routers
		vars            *sync.Map
		disableCompress bool
		idgen           func() string
		logs            *Logs
		closes          []func() error
		cache           cache.Driver
		contexts        []OnContextFunc
		exitContexts    []OnExitContextFunc
	}
)

// InternalNewServer 声明 [InternalServer]
//
// s 为实际的 [Server] 接口对象；
// requestIDKey 表示客户端提交的 X-Request-ID 报头名；
// problemPrefix 可以为空；
//
// NOTE: 此为内部使用函数，由调用者保证参数的正确性。
//
// NOTE: [Server] 的实现者，不应该重新实现 [InternalServer] 已经实现的接口，
// 否则可能出现 [InternalServer] 中的调用与 [Server] 的实现调用不同的问题。
// 比如重新实现了 [Server.Location]，那么将出现 [InternalServer] 内部的 Location
// 与 新实现的 Location 返回不同值的情况。
func InternalNewServer(s Server, name, ver string, loc *time.Location, logs *Logs, idgen func() string, l *locale.Locale, c cache.Driver, codec *Codec, requestIDKey, problemPrefix string, o ...RouterOption) *InternalServer {
	is := &InternalServer{
		server: s,

		name:    name,
		version: ver,

		locale:   l,
		location: loc,
		uptime:   time.Now().In(loc),

		requestIDKey: requestIDKey,
		codec:        codec,
		problems:     newProblems(problemPrefix),
		vars:         &sync.Map{},
		idgen:        idgen,
		logs:         logs,
		closes:       make([]func() error, 0, 10),
		cache:        c,
		contexts:     make([]OnContextFunc, 0, 10),
		exitContexts: make([]OnExitContextFunc, 0, 10),
	}
	is.initServices()
	is.routers = &Routers{
		g: group.NewOf(is.call,
			notFound,
			buildNodeHandle(http.StatusMethodNotAllowed),
			buildNodeHandle(http.StatusOK),
			o...),
	}

	return is
}

func (s *InternalServer) NewClient(c *http.Client, sel selector.Selector, m string, marshal func(any) ([]byte, error)) *Client {
	return NewClient(c, s.codec, sel, m, marshal, s.requestIDKey, s.server.UniqueID)
}

func (s *InternalServer) Config() *config.Config { return s.locale.Config() }

func (s *InternalServer) Locale() Locale { return s.locale }

func (s *InternalServer) Name() string { return s.name }

func (s *InternalServer) Version() string { return s.version }

func (s *InternalServer) Vars() *sync.Map { return s.vars }

func (s *InternalServer) CanCompress() bool { return !s.disableCompress }

func (s *InternalServer) SetCompress(enable bool) { s.disableCompress = !enable }

func (s *InternalServer) Location() *time.Location { return s.location }

func (s *InternalServer) Now() time.Time { return time.Now().In(s.Location()) }

func (s *InternalServer) Uptime() time.Time { return s.uptime }

func (s *InternalServer) Cache() cache.Cleanable { return s.cache }

func (s *InternalServer) ParseTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, s.Location())
}

func (s *InternalServer) UniqueID() string { return s.idgen() }

func (s *InternalServer) OnClose(f ...func() error) { s.closes = append(s.closes, f...) }

func (s *InternalServer) OnContext(f ...OnContextFunc) { s.contexts = append(s.contexts, f...) }

func (s *InternalServer) OnExitContext(f ...OnExitContextFunc) {
	s.exitContexts = append(s.exitContexts, f...)
}

func (s *InternalServer) Logs() *Logs { return s.logs }

func (s *InternalServer) Close() {
	slices.Reverse(s.closes)
	for _, f := range s.closes {
		if err := f(); err != nil { // 出错不退出，继续其它操作。
			s.Logs().ERROR().Error(err)
		}
	}
}

func (f PluginFunc) Plugin(s Server) { f(s) }

func (s *InternalServer) Use(p ...Plugin) {
	for _, pp := range p {
		pp.Plugin(s.server)
	}
}
