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
