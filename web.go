// SPDX-License-Identifier: MIT

// Package web 一个微型的 web 框架
//
// NOTE: 所有以 Internal 开头的公开函数，表示这个函数是仅模块可见的。
package web

import (
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v4"
	"golang.org/x/text/message"

	"github.com/issue9/web/server"
	"github.com/issue9/web/validation"
)

// Version 当前框架的版本
const Version = "0.55.0"

type (
	Server         = server.Server
	Context        = server.Context
	Options        = server.Options
	MiddlewareFunc = server.MiddlewareFunc
	Middleware     = server.Middleware
	HandlerFunc    = server.HandlerFunc
	Router         = server.Router
	Module         = server.Module
	Responser      = server.Responser
	Logger         = logs.Logger
	Rule           = validation.Rule
	Validator      = validation.Validator
	ValidateFunc   = validation.ValidateFunc

	// LocaleStringer 本地化字符串需要实在的接口
	//
	// 部分 error 返回可能也实现了该接口。
	LocaleStringer = localeutil.LocaleStringer
)

// NewServer 从 [Options] 初始化 [Server] 对象
func NewServer(name, version string, o *Options) (*Server, error) {
	return server.New(name, version, o)
}

// Phrase 生成本地化的语言片段
func Phrase(key message.Reference, v ...any) LocaleStringer { return localeutil.Phrase(key, v...) }

// NewRule 新建验证规则
func NewRule(v Validator, key message.Reference, val ...any) *Rule {
	return validation.NewRule(v, key, val...)
}
