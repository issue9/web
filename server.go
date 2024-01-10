// SPDX-License-Identifier: MIT

package web

import (
	"io/fs"
	"net/http"
	"sync"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/config"
	"github.com/issue9/mux/v7/types"
	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/selector"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

// Server 服务接口
type Server interface {
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
	// 与 [time.Now] 的区别在于 Now() 基于当前时区
	Now() time.Time

	// ParseTime 分析基于当前时区的时间
	ParseTime(layout, value string) (time.Time, error)

	// Serve 开始 HTTP 服务
	//
	// 这是个阻塞方法，会等待 [Server.Close] 执行完之后才返回。
	// 始终返回非空的错误对象，如果是被 [Server.Close] 关闭的，也将返回 [http.ErrServerClosed]。
	Serve() error

	// Close 触发关闭服务事件
	//
	// 需要等到 [Server.Serve] 返回才表示服务结束。
	// 调用此方法表示 [Server] 的生命周期结束，对象将处于不可用状态。
	Close(shutdownTimeout time.Duration)

	// OnClose 注册关闭服务时需要执行的函数
	//
	// NOTE: 按注册的相反顺序执行。
	OnClose(...func() error)

	// Config 当前项目配置文件的管理
	Config() *config.Config

	// Logs 提供日志接口
	Logs() *Logs

	// GetRouter 获取指定名称的路由
	GetRouter(name string) *Router

	// NewRouter 声明新路由
	NewRouter(name string, matcher RouterMatcher, o ...RouterOption) *Router

	// RemoveRouter 删除路由
	RemoveRouter(name string)

	// Routers 返回所有的路由
	Routers() []*Router

	// UseMiddleware 对所有的路由使用中间件
	UseMiddleware(...Middleware)

	// NewContext 从标准库的参数初始化 Context 对象
	//
	// 这适合从标准库的请求中创建 [web.Context] 对象。
	// [types.Route] 类型的参数需要用户通过 [types.NewContext] 自行创建。
	//
	// NOTE: 由此方法创建的对象在整个会话结束后会被回收.
	NewContext(http.ResponseWriter, *http.Request, types.Route) *Context

	// NewClient 基于当前对象的相关字段创建 [Client] 对象
	//
	// 功能与 [NewClient] 相同，缺少的参数直接采用 [Server] 关联的字段。
	NewClient(client *http.Client, s selector.Selector, marshalName string, marshal func(any) ([]byte, error)) *Client

	// SetCompress 设置压缩功能
	//
	// 在服务器性能吃紧的情况下可以采用此方法禁用压缩。
	//
	// NOTE: 仅对输出内容启作用，读取内容始终是按照提交的 Content-Encoding 指定算法进行解析。
	SetCompress(enable bool)

	// CanCompress 当前是否拥有压缩功能
	CanCompress() bool

	// Problems Problem 管理
	Problems() *Problems

	// Services 服务管理接口
	Services() *Services

	// Locale 提供本地化相关功能
	Locale() Locale
}

// Locale 提供与本地化相关的功能
type Locale interface {
	catalog.Catalog

	// ID 返回默认的语言标签
	ID() language.Tag

	// LoadMessages 从 fsys 中加载符合 glob 的本地化文件
	//
	// 根据 [Server.Config] 处理文件格式，如果文件格式不被 [Server.Config] 支持，将无法加载。
	LoadMessages(glob string, fsys ...fs.FS) error

	// Printer 最符合 [Locale.ID] 的 [message.Printer] 对象
	Printer() *message.Printer

	// Sprintf 等同于 Locale.Printer.Sprintf
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

// InternalServer 这是一个内部使用的类型，提供了部分 [Server] 接口的实现
type InternalServer struct {
	server       Server
	requestIDKey string
	codec        *Codec
	services     *Services
	problems     *Problems
}

// InternalNewServer 声明 [InternalServer]
//
// requestIDKey 表示客户端提交的 X-Request-ID 报头名，如果为空则采用 "X-Request-ID"；
// problemPrefix 如果为空会采用 [ProblemAboutBlank] 作为默认值；
func InternalNewServer(s Server, codec *Codec, requestIDKey, problemPrefix string) *InternalServer {
	if s == nil {
		panic("s 不能为空")
	}

	if requestIDKey == "" {
		requestIDKey = header.RequestIDKey
	}

	if problemPrefix == "" {
		problemPrefix = ProblemAboutBlank
	}

	return &InternalServer{
		server:       s,
		requestIDKey: requestIDKey,
		codec:        codec,
		services:     initServices(s),
		problems:     newProblems(problemPrefix),
	}
}

func (s *InternalServer) Codec() *Codec { return s.codec }

func (s *InternalServer) RequestIDKey() string { return s.requestIDKey }
