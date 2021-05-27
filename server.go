// SPDX-License-Identifier: MIT

package web

import (
	"io/fs"
	"net/http"

	"github.com/issue9/logs/v2"
	"golang.org/x/text/message/catalog"

	"github.com/issue9/web/config"
	"github.com/issue9/web/module"
	"github.com/issue9/web/result"
	"github.com/issue9/web/server"
)

type (
	Server      = server.Server
	Context     = server.Context
	Options     = server.Options
	Filter      = server.Filter
	Module      = server.Module
	Initializer = module.Initializer
	HandlerFunc = server.HandlerFunc
)

// LoadServer 从配置文件加载并实例化 Server 对象
func LoadServer(name, version string, f fs.FS, c catalog.Catalog, build result.BuildFunc) (*Server, error) {
	return config.NewServer(name, version, f, c, build)
}

// DefaultServer 返回一个采用默认值进初始化的 *Server 实例
func DefaultServer(name, version string) (*Server, error) {
	return NewServer(name, version, logs.New(), &Options{})
}

// NewServer 返回 *Server 实例
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
