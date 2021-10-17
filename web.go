// SPDX-License-Identifier: MIT

// Package web 一个微型的 RESTful API 框架
package web

import (
	"io/fs"
	"net/http"

	"github.com/issue9/localeutil"

	"github.com/issue9/web/config"
	"github.com/issue9/web/content"
	"github.com/issue9/web/serialization"
	"github.com/issue9/web/server"
)

// Version 当前框架的版本
const Version = "0.44.0"

type (
	Server       = server.Server
	Context      = server.Context
	Options      = server.Options
	Filter       = server.Filter
	Module       = server.Module
	HandlerFunc  = server.HandlerFunc
	Responser    = server.Responser
	Router       = server.Router
	Action       = server.Action
	ResultFields = content.ResultFields
	Files        = serialization.Files

	// LocaleStringer 本地化字符串需要实在的接口
	//
	// 部分 error 返回可能也实现了该接口。
	LocaleStringer = localeutil.LocaleStringer

	// OptionsFunc 用于对 Options 对象进行修改
	OptionsFunc func(*Options)
)

// LoadServer 从配置文件加载并实例化 Server 对象
//
// files 指定了用于加载本地化的方法，同时也用于加载配置文件；
// conf 用于指定项目的配置文件，根据扩展由 serialization.Files 负责在 f 查找文件加载；
// o 用于在初始化 Server 之前，加载配置文件之后，对 *Options 进行一次修改；
func LoadServer(name, version string, files *Files, f fs.FS, conf string, o OptionsFunc) (*Server, error) {
	opt, err := config.NewOptions(files, f, conf)
	if err != nil {
		return nil, err
	}
	if o != nil {
		o(opt)
	}
	return NewServer(name, version, opt)
}

// NewServer 返回 *Server 实例
func NewServer(name, version string, o *Options) (*Server, error) {
	return server.New(name, version, o)
}

// GetServer 从请求中获取 *Server 实例
func GetServer(r *http.Request) *Server { return server.GetServer(r) }

// NewContext 构建 *Context 实例
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return server.NewContext(w, r)
}

func Status(status int) Responser { return server.Status(status) }

func Object(status int, body interface{}, headers map[string]string) Responser {
	return server.Object(status, body, headers)
}

// Phrase 生成本地化的语言片段
func Phrase(key string, v ...interface{}) localeutil.LocaleStringer {
	return localeutil.Phrase(key, v...)
}

// Created 201
func Created(v interface{}, location string) Responser {
	if location == "" {
		return Object(http.StatusCreated, v, nil)
	}

	return Object(http.StatusCreated, v, map[string]string{
		"Location": location,
	})
}

// NotFound 404
func NotFound() Responser { return Status(http.StatusNotFound) }

// NoContent 204
func NoContent() Responser { return Status(http.StatusNoContent) }

// NotImplemented 501
func NotImplemented() Responser { return Status(http.StatusNotImplemented) }
