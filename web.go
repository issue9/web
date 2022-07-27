// SPDX-License-Identifier: MIT

// Package web 一个微型的 web 框架
package web

import (
	"github.com/issue9/localeutil"
	"github.com/issue9/logs/v4"
	"golang.org/x/text/message"

	"github.com/issue9/web/problem"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/response"
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
	Responser      = response.Responser
	Logger         = logs.Logger
	Rule           = problem.Rule

	// LocaleStringer 本地化字符串需要实在的接口
	//
	// 部分 error 返回可能也实现了该接口。
	LocaleStringer = localeutil.LocaleStringer
)

// NewServer 从 Options 初始化 Server 对象
func NewServer(name, version string, o *Options) (*Server, error) {
	return server.New(name, version, o)
}

// Phrase 生成本地化的语言片段
func Phrase(key message.Reference, v ...any) LocaleStringer { return localeutil.Phrase(key, v...) }

// NewRule 新建验证规则
func NewRule(v problem.Validator, key message.Reference, val ...any) *Rule {
	return problem.NewRule(v, key, val...)
}
