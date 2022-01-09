// SPDX-License-Identifier: MIT

package server

import (
	"errors"
	"io/fs"
	"net/http"

	"github.com/issue9/mux/v6"
	"github.com/issue9/mux/v6/group"
)

const defaultIndex = "index.html"

type Router = mux.RouterOf[HandlerFunc]

// MiddlewareFunc 中间件函数
type MiddlewareFunc = mux.MiddlewareFuncOf[HandlerFunc]

// AcceptMiddleware 提供限定 accept 的中间件
func AcceptMiddleware(ct ...string) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return Accept(next, ct...)
	}
}

// Accept 提供限定 accept 的中间件
func Accept(next HandlerFunc, ct ...string) HandlerFunc {
	return func(ctx *Context) Responser {
		for _, c := range ct {
			if c == ctx.OutputMimetypeName {
				return next(ctx)
			}
		}
		return Status(http.StatusNotAcceptable)
	}
}

// NewRouter 构建基于 matcher 匹配的路由操作实例
//
// domain 仅用于 URL 生成地址，并不会对路由本身产生影响，可以为空。
func (srv *Server) NewRouter(name, domain string, matcher group.Matcher, m ...MiddlewareFunc) *Router {
	return srv.group.New(name, matcher, m, mux.URLDomain(domain))
}

// Routes 返回所有路由的注册路由项
//
// 第一个键名表示路由名称，第二键值表示路由项地址，值表示该路由项支持的请求方法；
func (srv *Server) Routes() map[string]map[string][]string {
	return srv.group.Routes()
}

func (srv *Server) Routers() []*Router { return srv.group.Routers() }

func (srv *Server) Router(name string) *Router { return srv.group.Router(name) }

func (srv *Server) RemoveRouter(name string) { srv.group.Remove(name) }

// FileServer 提供静态文件服务
//
// fsys 为文件系统，如果为空则采用 srv.FS；
// name 表示参数名称；
// index 表示目录下的默认文件名；
func (srv *Server) FileServer(fsys fs.FS, name, index string) HandlerFunc {
	if fsys == nil {
		fsys = srv
	}

	if name == "" {
		panic("参数 name 不能为空")
	}

	if index == "" {
		index = defaultIndex
	}

	return func(ctx *Context) Responser {
		p, _ := ctx.params.Get(name) // 空值也是允许的值

		err := mux.ServeFile(fsys, p, index, ctx.Response, ctx.Request)
		switch {
		case errors.Is(err, fs.ErrPermission):
			return ctx.Error(http.StatusForbidden, err)
		case errors.Is(err, fs.ErrNotExist):
			return ctx.Error(http.StatusNotFound, err)
		case err != nil:
			return ctx.Error(http.StatusInternalServerError, err)
		default:
			return nil
		}
	}
}
