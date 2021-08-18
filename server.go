// SPDX-License-Identifier: MIT

package web

import (
	"io/fs"
	"net/http"

	"github.com/issue9/web/config"
	"github.com/issue9/web/content"
	"github.com/issue9/web/server"
)

type (
	Server      = server.Server
	Context     = server.Context
	Options     = server.Options
	Filter      = server.Filter
	Module      = server.Module
	HandlerFunc = server.HandlerFunc
	Responser   = server.Responser
	Router      = server.Router
)

// LoadServer 从配置文件加载并实例化 Server 对象
//
// logs 和 web 分别表示相对于 f 的日志和项目配置文件；
func LoadServer(name, version string, build content.BuildResultFunc, f fs.FS, logs, web string) (*Server, error) {
	return config.NewServer(name, version, build, f, logs, web)
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
