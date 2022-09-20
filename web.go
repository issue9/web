// SPDX-License-Identifier: MIT

// Package web 模块化的 web 框架
//
// NOTE: 所有以 Internal 开头的公开函数，表示这个函数是仅模块可见的。
package web

import (
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v4"
	"github.com/issue9/query/v3"
	"golang.org/x/text/message"

	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/server"
)

// Version 当前框架的版本
const Version = "0.59.1"

type (
	Server         = server.Server
	Context        = server.Context
	Options        = server.Options
	MiddlewareFunc = server.MiddlewareFunc
	Middleware     = server.Middleware
	HandlerFunc    = server.HandlerFunc
	Router         = server.Router
	Responser      = server.Responser
	ResponserFunc  = server.ResponserFunc
	CTXSanitizer   = server.CTXSanitizer
	Rule           = server.Rule
	Validation     = server.Validation
	Validator      = server.Validator
	ValidateFunc   = server.ValidateFunc
	Logger         = logs.Logger

	// QueryUnmarshaler 对查询参数的解析接口
	QueryUnmarshaler = query.Unmarshaler

	// LocaleStringer 本地化字符串需要实在的接口
	//
	// 部分 error 返回可能也实现了该接口。
	LocaleStringer = localeutil.LocaleStringer
)

func NewServer(name, version string, o *Options) (*Server, error) {
	return server.New(name, version, o)
}

// Phrase 生成本地化的语言片段
func Phrase(key message.Reference, v ...any) LocaleStringer { return localeutil.Phrase(key, v...) }

// NewRule 新建验证规则
func NewRule(msg LocaleStringer, v Validator) *Rule { return server.NewRule(msg, v) }

// NewRuleFunc 新建验证规则
func NewRuleFunc(msg LocaleStringer, f func(any) bool) *Rule {
	return server.NewRuleFunc(msg, f)
}

// StackError 为 err 带上调用信息
//
// 位置从调用 StackError 开始。
// 如果 err 为 nil，则返回 nil，如果 err 本身就为 StackError 返回的类型，则原样返回。
//
// 如果需要输出调用堆栈信息，需要指定 %+v 标记。
func StackError(err error) error { return errs.StackError(err) }

// Errors 合并多个非空错误为一个错误
//
// 有关 Is 和 As 均是按顺序找第一个返回 true 的即返回。
func Errors(err ...error) error { return errs.Errors(err...) }
