// SPDX-License-Identifier: MIT

package web

import (
	"context"
	"io/fs"
	"net/http"
	"sync"
	"time"

	"github.com/issue9/config"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/cache"
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
	Serve() (err error)

	// Close 触发关闭服务事件
	//
	// 需要等到 [Server.Serve] 返回才表示服务结束。
	// 调用此方法表示 [Server] 的生命周期结束，对象将处于不可用状态。
	Close(shutdownTimeout time.Duration)

	// OnClose 注册关闭服务时需要执行的函数
	//
	// NOTE: 按注册的相反顺序执行。
	OnClose(f ...func() error)

	// Config 当前项目配置文件的管理
	Config() *config.Config

	// Logs 日志接口
	Logs() Logs

	// GetRouter 获取指定名称的路由
	GetRouter(name string) *Router

	// NewRouter 声明新路由
	NewRouter(name string, matcher RouterMatcher, o ...RouterOption) *Router

	// RemoveRouter 删除路由
	RemoveRouter(name string)

	// Routers 返回所有的路由
	Routers() []*Router

	// UseMiddleware 对所有的路由使用中间件
	UseMiddleware(m ...Middleware)

	// NewContext 从标准库的参数初始化 Context 对象
	//
	// NOTE: 这适合从标准库的请求中创建 [web.Context] 对象，
	// 但是部分功能会缺失，比如地址中的参数信息，以及 [Context.Route] 等。
	NewContext(w http.ResponseWriter, r *http.Request) *Context

	// NewClient 基于当前对象的 [Server.Codec] 初始化 [Client]
	NewClient(client *http.Client, selector Selector, marshalName string, marshal func(any) ([]byte, error)) *Client

	// SetCompress 设置压缩功能
	//
	// 在服务器性能吃紧的情况下可以采用此方法禁用压缩。
	//
	// NOTE: 仅对输出内容启作用，读取内容始终是按照提交的 Content-Encoding 指定算法进行解析。
	SetCompress(enable bool)

	// CanCompress 当前是否拥有压缩功能
	CanCompress() bool

	// Problems Problem 管理
	Problems() Problems

	// Services 服务管理接口
	Services() Services

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

	// NewPrinter 声明最符合 tag 的 [message.Printer] 对象
	//
	// NOTE: 随着 [Locale.SetString] 的不断调用，相同参数返回的对象可能会变化。
	NewPrinter(tag language.Tag) *message.Printer

	// SetString 添加新的翻译项
	SetString(tag language.Tag, key, msg string) error

	// SetMacro 添加新的翻译项
	SetMacro(tag language.Tag, name string, msg ...catalog.Message) error

	// Set 添加新的翻译项
	Set(tag language.Tag, key string, msg ...catalog.Message) error
}

// Problems Problem 管理接口
type Problems interface {
	// Add 添加新项
	//
	// NOTE: 已添加的内容无法修改，如果确实有需求，
	// 只能通过修改翻译项的方式间接进行修改。
	Add(status int, p ...LocaleProblem) Problems

	// Init 将指定 ID 的项注入 [RFC7807] 对象
	Init(problem *RFC7807, id string, p *message.Printer)

	// Visit 遍历错误代码
	//
	// visit 签名：
	//
	//	func(id string, status int, title, detail LocaleStringer)
	//
	// id 为错误代码，不包含前缀部分；
	// status 该错误代码反馈给用户的 HTTP 状态码；
	// title 错误代码的简要描述；
	// detail 错误代码的明细；
	Visit(visit func(id string, status int, title, detail LocaleStringer))

	// Prefix 所有 ID 的统一前缀
	Prefix() string
}

type LocaleProblem struct {
	ID            string
	Title, Detail LocaleStringer
}

// Services 服务管理接口
type Services interface {
	// Add 添加并运行新的服务
	//
	// title 是对该服务的简要说明；
	Add(title LocaleStringer, f Service)

	// AddFunc 将函数 f 作为服务添加并运行
	AddFunc(title LocaleStringer, f func(context.Context) error)

	// AddCron 添加新的定时任务
	//
	// title 是对该服务的简要说明；
	// spec cron 表达式，支持秒；
	// delay 是否在任务执行完之后，才计算下一次的执行时间点。
	//
	// NOTE: 此功能依赖 [Server.UniqueID]。
	AddCron(title LocaleStringer, f JobFunc, spec string, delay bool)

	// AddAt 添加在某个时间点执行的任务
	//
	// title 是对该服务的简要说明；
	// at 指定的时间点；
	// delay 是否在任务执行完之后，才计算下一次的执行时间点。
	//
	// NOTE: 此功能依赖 [Server.UniqueID]。
	AddAt(title LocaleStringer, job JobFunc, at time.Time, delay bool)

	// AddJob 添加新的计划任务
	//
	// title 是对该服务的简要说明；
	// scheduler 计划任务的时间调度算法实现；
	// delay 是否在任务执行完之后，才计算下一次的执行时间点。
	//
	// NOTE: 此功能依赖 [Server.UniqueID]。
	AddJob(title LocaleStringer, job JobFunc, scheduler Scheduler, delay bool)

	// AddTicker 添加新的定时任务
	//
	// title 是对该服务的简要说明；
	// dur 时间间隔；
	// imm 是否立即执行一次该任务；
	// delay 是否在任务执行完之后，才计算下一次的执行时间点。
	//
	// NOTE: 此功能依赖 [Server.UniqueID]。
	AddTicker(title LocaleStringer, job JobFunc, dur time.Duration, imm, delay bool)

	// Visit 访问所有的服务
	//
	// visit 的原型为：
	//  func(title LocaleStringer, state State, err error)
	//
	// title 为服务的说明；
	// state 为服务的当前状态；
	// err 只在 state 为 [Failed] 时才有的错误说明；
	Visit(visit func(title LocaleStringer, state State, err error))

	// VisitJobs 访问所有的计划任务
	//
	// visit 原型为：
	//  func(title LocaleStringer, prev, next time.Time, state State, delay bool, err error)
	// title 为计划任务的说明；
	// prev 和 next 表示任务的上一次执行时间和下一次执行时间；
	// state 表示当前的状态；
	// delay 表示该任务是否是执行完才开始计算下一次任务时间的；
	// err 表示这个任务的出错状态；
	VisitJobs(visit func(LocaleStringer, time.Time, time.Time, State, bool, error))
}
