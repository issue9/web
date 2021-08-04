// SPDX-License-Identifier: MIT

package web

import (
	"io/fs"
	"net/http"

	"github.com/issue9/logs/v3"

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
)

// LoadServer 从配置文件加载并实例化 Server 对象
func LoadServer(name, version string, f fs.FS, build content.BuildResultFunc) (*Server, error) {
	return config.NewServer(name, version, f, build)
}

// DefaultServer 返回一个采用默认值进初始化的 *Server 实例
func DefaultServer(name, version string) (*Server, error) {
	l, err := logs.New(nil)
	if err != nil {
		return nil, err
	}
	return NewServer(name, version, l, nil)
}

// NewServer 返回 *Server 实例
func NewServer(name, version string, logs *logs.Logs, o *Options) (*Server, error) {
	return server.New(name, version, logs, o)
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
