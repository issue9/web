// SPDX-FileCopyrightText: 2018-2025 caixw
//
// SPDX-License-Identifier: MIT

//go:generate web locale -l=und -m -f=yaml ./
//go:generate web update-locale -src=./locales/und.yaml -dest=./locales/zh.yaml
//go:generate go run ./make_problems.go

// Package web 通用的 web 开发框架
//
// NOTE: 所有以 Internal 开头的函数和对象都是模块内部使用的。
package web

import (
	"io"
	"runtime/debug"

	"github.com/issue9/cache"
	"github.com/issue9/config"
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v7"
	"github.com/issue9/query/v3"

	"github.com/issue9/web/internal/errs"
)

// Version 当前框架的版本
const Version = "0.104.0"

type (
	Logger   = logs.Logger
	Logs     = logs.Logs
	AttrLogs = logs.AttrLogs

	// Cache 缓存内容的访问接口
	Cache = cache.Cache

	// FieldError 表示配置文件中的字段错误
	FieldError = config.FieldError

	// QueryUnmarshaler 对查询参数的解析接口
	QueryUnmarshaler = query.Unmarshaler

	// LocaleStringer 本地化字符串需要实现的接口
	LocaleStringer = localeutil.Stringer

	StringPhrase = localeutil.StringPhrase

	// MarshalFunc 序列化函数原型
	//
	// NOTE: MarshalFunc 可根据需求自行决定 [Problem] 和 [server.RenderResponse] 的输出格式。
	//
	// NOTE: MarshalFunc 的作用是输出内容，所以在实现中不能调用 [Context.Render] 等输出方法。
	//
	// NOTE: 不采用流的方式处理数据的原因是因为：编码过程中可能会出错，
	// 此时需要修改状态码，流式的因为有内容输出，状态码也已经固定，无法修改。
	MarshalFunc func(*Context, any) ([]byte, error)

	// UnmarshalFunc 反序列化函数原型
	//
	// NOTE: 参数 [io.Reader] 必定不会为空。
	UnmarshalFunc func(io.Reader, any) error
)

// GetAppVersion 获得应用的版本号
//
// 如果当前应用没有指定版本号，则采用参数 v 作为返回值。
//
// NOTE: 只有 go1.24 及之后且 -buildinfo 参数不为 false 编译的程序才会带版本信息，否则始终返回 v。
func GetAppVersion(v string) string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.Main.Version
	}
	return v
}

// NewCache 声明带有统一前缀的缓存接口
func NewCache(prefix string, c Cache) Cache { return cache.Prefix(c, prefix) }

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
func NewFieldError(field string, msg any) *FieldError { return config.NewFieldError(field, msg) }

// NewLocaleError 本地化的错误信息
func NewLocaleError(format string, v ...any) error { return localeutil.Error(format, v...) }

// NewError 用 HTTP 状态码包装一个错误信息
//
// status 表示 HTTP 状态码；
// err 被包装的错误信息，如果是空值，将会 panic；
//
// 此方法返回的错误，在 [Context.Error] 和 [WithRecovery] 中会被识别且按指定的状态码输出。
func NewError(status int, err error) error { return errs.NewError(status, err) }

// SprintError 将 err 转换为字符串
//
// p 如果 err 实现了 [LocaleStringer]，将采用 p 进行转换；
// detail 是否输出调用堆栈；
func SprintError(p *localeutil.Printer, detail bool, err error) string {
	return errs.Sprint(p, err, detail)
}
