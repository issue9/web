// SPDX-License-Identifier: MIT

// Package web 一个微型的 web 框架
package web

import (
	"net/http"
	"strconv"
	"time"

	"github.com/issue9/localeutil"

	"github.com/issue9/web/server"
)

// Version 当前框架的版本
const Version = "0.48.0"

type (
	Server         = server.Server
	Context        = server.Context
	Options        = server.Options
	MiddlewareFunc = server.MiddlewareFunc
	Middleware     = server.Middleware
	HandlerFunc    = server.HandlerFunc
	Responser      = server.Responser
	Router         = server.Router
	ResultFields   = server.ResultFields
	Result         = server.Result
	Module         = server.Module

	// LocaleStringer 本地化字符串需要实在的接口
	//
	// 部分 error 返回可能也实现了该接口。
	LocaleStringer = localeutil.LocaleStringer
)

// NewServer 从 Options 初始化 Server 对象
func NewServer(name, version string, o *Options) (*Server, error) {
	return server.New(name, version, o)
}

func Status(status int) *Responser { return server.Status(status) }

// Phrase 生成本地化的语言片段
func Phrase(key string, v ...any) LocaleStringer {
	return localeutil.Phrase(key, v...)
}

func Created(v any, location string) *Responser {
	resp := Status(http.StatusCreated).Body(v)
	if location != "" {
		resp.SetHeader("Location", location)
	}
	return resp
}

// OK 返回 200 状态码下的对象
func OK(v any) *Responser { return Status(http.StatusOK).Body(v) }

func NotFound() *Responser { return Status(http.StatusNotFound) }

func NoContent() *Responser { return Status(http.StatusNoContent) }

func NotImplemented() *Responser { return Status(http.StatusNotImplemented) }

// RetryAfter 返回 Retry-After 报头内容
//
// 一般适用于 301 和 503 报文。
//
// status 表示返回的状态码；seconds 表示秒数，如果想定义为时间格式，
// 可以采用 RetryAt 函数，两个功能是相同的，仅是时间格式上有差别。
func RetryAfter(status int, seconds uint64) *Responser {
	return Status(status).SetHeader("Retry-After", strconv.FormatUint(seconds, 10))
}

func RetryAt(status int, at time.Time) *Responser {
	return Status(status).SetHeader("Retry-After", at.UTC().Format(http.TimeFormat))
}
