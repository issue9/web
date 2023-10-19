// SPDX-License-Identifier: MIT

//go:generate web locale -l=und -m -f=yaml ./
//go:generate web update-locale -src=./locales/und.yaml -dest=./locales/zh-CN.yaml
//go:generate go run ./make_problems.go

// Package web 通用的 web 开发框架
package web

import (
	"context"
	"io/fs"

	"github.com/issue9/config"
	"github.com/issue9/localeutil"
	"github.com/issue9/mux/v7"
	"github.com/issue9/mux/v7/group"
	"github.com/issue9/mux/v7/types"
	"github.com/issue9/query/v3"
	"github.com/issue9/scheduled"

	"github.com/issue9/web/cache"
	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/logs"
)

// Version 当前框架的版本
const Version = "0.86.0"

// 服务的几种状态
const (
	Stopped = scheduled.Stopped // 停止状态，默认状态
	Running = scheduled.Running // 正在运行
	Failed  = scheduled.Failed  // 出错，不再执行后续操作
)

var errUnsupportedSerialization = NewLocaleError("unsupported serialization")

type (
	Logger = logs.Logger
	Logs   = logs.Logs

	// Cache 缓存内容的访问接口
	Cache = cache.Cache

	// FieldError 表示配置文件中的字段错误
	FieldError = config.FieldError

	// QueryUnmarshaler 对查询参数的解析接口
	QueryUnmarshaler = query.Unmarshaler

	// LocaleStringer 本地化字符串需要实在的接口
	LocaleStringer = localeutil.Stringer

	StringPhrase = localeutil.StringPhrase

	// Service 长期运行的服务需要实现的接口
	Service interface {
		// Serve 运行服务
		//
		// 这是个阻塞方法，实现者需要正确处理 [context.Context.Done] 事件。
		// 如果是通过 [context.Context] 的相关操作取消的，应该返回 [context.Context.Err]。
		Serve(context.Context) error
	}

	ServiceFunc func(context.Context) error

	// State 服务状态
	//
	// 以下设置用于 restdoc
	//
	// @type string
	// @enum stopped running failed
	State = scheduled.State

	JobFunc       = scheduled.JobFunc
	Scheduler     = scheduled.Scheduler
	SchedulerFunc = scheduled.SchedulerFunc

	Router         = mux.RouterOf[HandlerFunc]
	Prefix         = mux.PrefixOf[HandlerFunc]
	Resource       = mux.ResourceOf[HandlerFunc]
	RouterMatcher  = group.Matcher
	RouterOption   = mux.Option
	MiddlewareFunc = types.MiddlewareFuncOf[HandlerFunc]
	Middleware     = types.MiddlewareOf[HandlerFunc]

	// HandlerFunc 路由的处理函数原型
	//
	// 向客户端输出内容的有两种方法，一种是通过 [Context.Write] 方法；
	// 或是返回 [Responser] 对象。前者在调用 [Context.Write] 时即输出内容，
	// 后者会在整个请求退出时才将 [Responser] 进行编码输出。
	//
	// 返回值可以为空，表示在中间件执行过程中已经向客户端输出同内容。
	HandlerFunc func(*Context) Responser

	// BuildMarshalFunc 生成特定于 [Context] 的 [MarshalFunc]
	//
	// 如果传递的参数是空值，应该返回一个默认的 [MarshalFunc] 实现，
	// 该实现将被用于 [Client] 的相关功能。
	BuildMarshalFunc = func(*Context) MarshalFunc

	// MarshalFunc 序列化函数原型
	//
	// NOTE: MarshalFunc 的作用是输出内容，所以在实现中不能调用 [Context.Render] 等输出方法。
	MarshalFunc = func(any) ([]byte, error)

	// UnmarshalFunc 反序列化函数原型
	UnmarshalFunc = func([]byte, any) error
)

// ErrUnsupportedSerialization 返回不支持序列化的错误信息
func ErrUnsupportedSerialization() error { return errUnsupportedSerialization }

func (f ServiceFunc) Serve(ctx context.Context) error { return f(ctx) }

// Phrase 生成本地化的语言片段
func Phrase(key string, v ...any) LocaleStringer { return localeutil.Phrase(key, v...) }

// NewStackError 为 err 带上调用信息
//
// 位置从调用 NewStackError 开始。如果 err 为 nil，则返回 nil。
// 多次调用 NewStackError 包装，则返回第一次包装的返回值。
func NewStackError(err error) error { return errs.NewDepthStackError(2, err) }

// NewFieldError 返回表示配置文件错误的对象
//
// field 表示错误的字段名；
// msg 表示错误信息，可以是任意类型，如果 msg 是 [FieldError] 类型，
// 那么此操作相当于调用了 [FieldError.AddFieldParent]；
func NewFieldError(field string, msg any) *FieldError {
	return config.NewFieldError(field, msg)
}

// NewLocaleError 本地化的错误信息
func NewLocaleError(format string, v ...any) error {
	return localeutil.Error(format, v...)
}

// NewError 用 HTTP 状态码包装一个错误信息
//
// status 表示 HTTP 状态码；
// err 被包装的错误信息，如果所有错误都是空值，将会 panic；
//
// 此方法返回的错误，在 [Context.Error] 会被识别且按指定的状态码输出。
func NewError(status int, err error) error { return errs.NewError(status, err) }

// FileServer 构建静态文件服务对象
//
// fsys 为文件系统；
// name 表示地址中表示文件名部分的参数名称；
// index 表示目录下的默认文件名；
func FileServer(fsys fs.FS, name, index string) HandlerFunc {
	if name == "" {
		panic("参数 name 不能为空")
	}

	return func(ctx *Context) Responser {
		p, _ := ctx.Route().Params().Get(name) // 空值也是允许的值
		if err := mux.ServeFile(fsys, p, index, ctx, ctx.Request()); err != nil {
			return ctx.Error(err, "")
		}
		return nil
	}
}
