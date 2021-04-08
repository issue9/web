// SPDX-License-Identifier: MIT

package web

import (
	"net/http"

	"github.com/issue9/logs/v2"

	"github.com/issue9/web/server"
)

type (
	Server      = server.Server
	Context     = server.Context
	Options     = server.Options
	Filter      = server.Filter
	Module      = server.Module
	Tag         = server.Tag
	HandlerFunc = server.HandlerFunc
)

// New 返回 *Server 实例
func NewServer(name, version string, logs *logs.Logs, o *Options) (*Server, error) {
	return server.New(name, version, logs, o)
}

// GetServer 从请求中获取 *Server 实例
func GetServer(r *http.Request) *Server {
	return server.GetServer(r)
}

// NewContext 构建 *Context 实例
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return server.NewContext(w, r)
}

// NewModule 声明一个新的模块
func NewModule(id, desc string, deps ...string) *Module {
	return server.NewModule(id, desc, deps...)
}
