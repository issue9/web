// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

//go:generate gomsgs extract -l=und -m -f=yaml -tag=comment -func=github.com/issue9/localeutil.Phrase,github.com/issue9/localeutil.Error,github.com/issue9/localeutil.StringPhrase,github.com/issue9/web.Phrase,github.com/issue9/web.StringPhrase,github.com/issue9/web.NewError,github.com/issue9/web.Context.Sprintf,github.com/issue9/web.Locale.Sprintf,github.com/issue9/web/openapi.Document.ParameterizedDoc ./
//go:generate gomsgs update -src=./locales/und.yaml -dest=./locales/zh.yaml

// Package web 通用的 web 开发框架
package web

import (
	"github.com/issue9/cache"
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v7"
	"github.com/issue9/query/v3"

	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/negotiation"
)

// Version 当前框架的版本号
const Version = "1.0.0-beta.1"

type (
	Logger   = logs.Logger
	Logs     = logs.Logs
	AttrLogs = logs.AttrLogs

	// Cache 缓存内容的访问接口
	Cache = cache.Cache

	// QueryUnmarshaler 对查询参数的解析接口
	QueryUnmarshaler = query.Unmarshaler

	// LocaleStringer 本地化字符串需要实现的接口
	LocaleStringer = localeutil.Stringer

	StringPhrase = localeutil.StringPhrase

	Negotiation = negotiation.Negotiation[*Context]

	MediaType = negotiation.MediaType[*Context]

	Codec = negotiation.Codec[*Context]
)

// NewCache 声明带有统一前缀的缓存接口
func NewCache(prefix string, c Cache) Cache { return cache.Prefix(c, prefix) }

// Phrase 生成本地化的语言片段
func Phrase(key string, v ...any) LocaleStringer { return localeutil.Phrase(key, v...) }

// NewStackError 为 err 带上调用信息
//
// 位置从调用 NewStackError 开始。如果 err 为 nil，则返回 nil。
// 多次调用 NewStackError 包装，则返回第一次包装的返回值。
func NewStackError(err error) error { return errs.NewDepthStackError(2, err) }

// NewError 本地化的错误信息
func NewError(format string, v ...any) error { return localeutil.Error(format, v...) }

// SprintError 将 err 转换为字符串
//
// p 如果 err 实现了 [LocaleStringer]，将采用 p 进行转换；
// detail 是否输出调用堆栈；
func SprintError(p *localeutil.Printer, detail bool, err error) string {
	return errs.Sprint(p, err, detail)
}
