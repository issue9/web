// SPDX-License-Identifier: MIT

package server

import (
	"errors"
	"io/fs"
	"net/http"

	"github.com/issue9/mux/v6"
	"github.com/issue9/mux/v6/muxutil"
)

type (
	Router         = mux.RouterOf[HandlerFunc]
	Routers        = mux.RoutersOf[HandlerFunc]
	RouterOptions  = mux.OptionsOf[HandlerFunc]
	MiddlewareFunc = mux.MiddlewareFuncOf[HandlerFunc]
	Middleware     = mux.MiddlewareOf[HandlerFunc]
)

// Routers 路由管理接口
func (srv *Server) Routers() *Routers { return srv.routers }

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

	return func(ctx *Context) Responser {
		p, _ := ctx.params.Get(name) // 空值也是允许的值

		err := muxutil.ServeFile(fsys, p, index, ctx, ctx.Request())
		switch {
		case errors.Is(err, fs.ErrPermission):
			return Status(http.StatusForbidden)
		case errors.Is(err, fs.ErrNotExist):
			return Status(http.StatusNotFound)
		case err != nil:
			return ctx.Error(http.StatusInternalServerError, err)
		default:
			return nil
		}
	}
}
