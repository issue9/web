// SPDX-License-Identifier: MIT

// Package web 一个微型的 RESTful API 框架
package web

import (
	"io/fs"
	"net/http"

	"github.com/issue9/web/config"
	"github.com/issue9/web/content"
	"github.com/issue9/web/serialization"
	"github.com/issue9/web/server"
)

// Version 当前框架的版本
const Version = "0.42.0"

type (
	Server       = server.Server
	Context      = server.Context
	Options      = server.Options
	Filter       = server.Filter
	Module       = server.Module
	HandlerFunc  = server.HandlerFunc
	Responser    = server.Responser
	Router       = server.Router
	Tag          = server.Tag
	ResultFields = content.ResultFields
	Locale       = serialization.Locale
)

// LoadServer 从配置文件加载并实例化 Server 对象
//
// locale 指定了用于加载本地化的方法，同时其关联的 serialization.Files 也用于加载配置文件；
// logs 和 web 用于指定日志和项目的配置文件，根据扩展由 serialization.Files 负责在 f 查找文件加载；
func LoadServer(name, version string, build content.BuildResultFunc, l *Locale, f fs.FS, logs, web string) (*Server, error) {
	o, err := config.NewOptions(build, l, f, logs, web)
	if err != nil {
		return nil, err
	}
	return NewServer(name, version, o)
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
