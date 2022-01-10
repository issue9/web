// SPDX-License-Identifier: MIT

// Package web 一个微型的 web 框架
package web

import (
	"net/http"

	"github.com/issue9/localeutil"

	"github.com/issue9/web/server"
)

// Version 当前框架的版本
const Version = "0.46.0"

type (
	Server         = server.Server
	Context        = server.Context
	Options        = server.Options
	MiddlewareFunc = server.MiddlewareFunc
	HandlerFunc    = server.HandlerFunc
	Responser      = server.Responser
	Router         = server.Router
	ResultFields   = server.ResultFields
	Result         = server.Result

	// LocaleStringer 本地化字符串需要实在的接口
	//
	// 部分 error 返回可能也实现了该接口。
	LocaleStringer = localeutil.LocaleStringer
)

// NewServer 从 Options 初始化 Server 对象
func NewServer(name, version string, o *Options) (*Server, error) {
	return server.New(name, version, o)
}

func Status(status int) Responser { return server.Status(status) }

func Object(status int, body interface{}, headers map[string]string) Responser {
	return server.Object(status, body, headers)
}

// Phrase 生成本地化的语言片段
func Phrase(key string, v ...interface{}) LocaleStringer {
	return localeutil.Phrase(key, v...)
}

func Created(v interface{}, location string) Responser {
	if location == "" {
		return Object(http.StatusCreated, v, nil)
	}

	return Object(http.StatusCreated, v, map[string]string{
		"Location": location,
	})
}

// OK 返回 200 状态码下的对象
func OK(v interface{}) Responser { return Object(http.StatusOK, v, nil) }

func NotFound() Responser { return Status(http.StatusNotFound) }

func NoContent() Responser { return Status(http.StatusNoContent) }

func NotImplemented() Responser { return Status(http.StatusNotImplemented) }
