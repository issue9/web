// SPDX-License-Identifier: MIT

package server

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/issue9/web/content"
)

// Context 是对当次 HTTP 请求内容的封装
type Context struct {
	*content.Context
	server *Server

	// 与当前对话相关的时区
	Location *time.Location

	// 保存 Context 在存续期间的可复用变量
	//
	// 这是比 context.Value 更经济的传递变量方式。
	//
	// 如果仅需要在多个请求中传递参数，可直接使用 Server.Vars。
	Vars map[interface{}]interface{}
}

// Filter 针对 Context 的中间件
//
// Filter 和 github.com/issue9/mux.MiddlewareFunc 本质上没有任何区别，
// mux.MiddlewareFunc 更加的通用，可以复用市面上的大部分中间件，
// Filter 则更加灵活一些，适合针对当前框架新的中间件。
//
// 如果想要使用 mux.MiddlewareFunc，可以调用 Server.MuxGroups().Middlewares() 方法。
type Filter func(HandlerFunc) HandlerFunc

// ApplyFilters 将过滤器应用于处理函数 next
func ApplyFilters(next HandlerFunc, filter ...Filter) HandlerFunc {
	if l := len(filter); l > 0 {
		for i := l - 1; i >= 0; i-- {
			next = filter[i](next)
		}
	}
	return next
}

// NewContext 构建 *Context 实例
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	if ctx := r.Context().Value(contextKeyContext); ctx != nil {
		return ctx.(*Context)
	}
	return GetServer(r).NewContext(w, r)
}

// NewContext 构建 *Context 实例
//
//
// 如果不合规则，会以指定的状码退出。
// 比如 Accept 的内容与当前配置无法匹配，则退出(panic)并输出 NotAcceptable 状态码。
func (srv *Server) NewContext(w http.ResponseWriter, r *http.Request) *Context {
	ctx, status := srv.Content().NewContext(srv.logs.DEBUG(), w, r)
	if status != 0 {
		srv.errorHandlers.Exit(w, status)
	}

	c := &Context{
		server:   srv,
		Context:  ctx,
		Location: srv.Location(),
		Vars:     map[interface{}]interface{}{},
	}

	c.Request = r.WithContext(context.WithValue(r.Context(), contextKeyContext, c))

	return c
}

// Read 从客户端读取数据并转换成 v 对象
//
// 功能与 Unmarshal() 相同，只不过 Read() 在出错时，
// 会直接调用 Error() 处理：输出 422 的状态码，
// 并返回一个 false，告知用户转换失败。
// 如果是数据类型验证失败，则会输出以 code 作为错误代码的错误信息，
// 并返回 false，作为执行失败的通知。
func (ctx *Context) Read(v interface{}, code int) (ok bool) {
	if err := ctx.Unmarshal(v); err != nil {
		ctx.Error(http.StatusUnprocessableEntity, err)
	}

	if vv, ok := v.(CTXSanitizer); ok {
		if errors := vv.CTXSanitize(ctx); len(errors) > 0 {
			ctx.NewResultWithFields(code, errors).Render()
			return false
		}
	}

	return true
}

// Render 将 v 渲染给客户端
//
// 功能与 Marshal() 相同，只不过 Render() 在出错时，
// 会直接调用 Error() 处理，输出 500 的状态码。
//
// 如果需要具体控制出错后的处理方式，可以使用 Marshal 函数。
func (ctx *Context) Render(status int, v interface{}, headers map[string]string) {
	if err := ctx.Marshal(status, v, headers); err != nil {
		ctx.Error(http.StatusInternalServerError, err)
	}
}

// ClientIP 返回客户端的 IP 地址
//
// NOTE: 包含了端口部分。
//
// 获取顺序如下：
//  - X-Forwarded-For 的第一个元素
//  - Remote-Addr 报头
//  - X-Read-IP 报头
func (ctx *Context) ClientIP() string {
	ip := ctx.Request.Header.Get("X-Forwarded-For")
	if index := strings.IndexByte(ip, ','); index > 0 {
		ip = ip[:index]
	}
	if ip == "" && ctx.Request.RemoteAddr != "" {
		ip = ctx.Request.RemoteAddr
	}
	if ip == "" {
		ip = ctx.Request.Header.Get("X-Real-IP")
	}

	return strings.TrimSpace(ip)
}

// ServeFile 提供文件下载
//
// 文件可能提供连续的下载功能，其状态码是未定的，
// 所以提供了一个类似于 Render 的变体专门用于下载功能。
func (ctx *Context) ServeFile(p, index string, headers map[string]string) {
	dir := filepath.ToSlash(filepath.Dir(p))
	base := filepath.ToSlash(filepath.Base(p))
	ctx.ServeFileFS(os.DirFS(dir), base, index, headers)
}

// ServeFileFS 提供基于 fs.FS 的文件下载服
func (ctx *Context) ServeFileFS(f fs.FS, p, index string, headers map[string]string) {
	err := ctx.ServeFS(f, p, index, headers)

	switch {
	case errors.Is(err, fs.ErrPermission):
		ctx.Exit(http.StatusForbidden)
	case errors.Is(err, fs.ErrNotExist):
		ctx.NotFound()
	case err != nil:
		ctx.Error(http.StatusInternalServerError, err)
	}
}

// Created 201
func (ctx *Context) Created(v interface{}, location string) {
	if location == "" {
		ctx.Render(http.StatusCreated, v, nil)
	} else {
		ctx.Render(http.StatusCreated, v, map[string]string{
			"Location": location,
		})
	}
}

// NoContent 204
func (ctx *Context) NoContent() { ctx.Response.WriteHeader(http.StatusNoContent) }

// ResetContent 205
func (ctx *Context) ResetContent() { ctx.Response.WriteHeader(http.StatusResetContent) }

// NotFound 404
func (ctx *Context) NotFound() { ctx.Response.WriteHeader(http.StatusNotFound) }

// NotImplemented 501
func (ctx *Context) NotImplemented() { ctx.Response.WriteHeader(http.StatusNotImplemented) }
