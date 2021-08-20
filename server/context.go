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

// CTXSanitizer 提供对数据的验证和修正
//
// 但凡对象实现了该接口，那么在 Context.Read 和 Queries.Object
// 中会在解析数据成功之后，调用该接口进行数据验证。
type CTXSanitizer interface {
	CTXSanitize(*Context) content.Fields
}

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
	// 如果需要在多个请求中传递参数，可直接使用 Server.Vars。
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
// 如果不合规则，会以指定的状码退出。
// 比如 Accept 的内容与当前配置无法匹配，则退出(panic)并输出 NotAcceptable 状态码。
func (srv *Server) NewContext(w http.ResponseWriter, r *http.Request) *Context {
	ctx, status := srv.content.NewContext(srv.logs.DEBUG(), w, r)
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
		resp := ctx.Error(http.StatusUnprocessableEntity, err)
		ctx.renderResponser(resp)
		return false
	}

	if vv, ok := v.(CTXSanitizer); ok {
		if errors := vv.CTXSanitize(ctx); len(errors) > 0 {
			resp := ctx.Result(code, errors)
			ctx.renderResponser(resp)
			return false
		}
	}

	return true
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
func (ctx *Context) ServeFile(p, index string, headers map[string]string) Responser {
	dir := filepath.ToSlash(filepath.Dir(p))
	base := filepath.ToSlash(filepath.Base(p))
	return ctx.ServeFileFS(os.DirFS(dir), base, index, headers)
}

// ServeFileFS 提供基于 fs.FS 的文件下载服
func (ctx *Context) ServeFileFS(f fs.FS, p, index string, headers map[string]string) Responser {
	err := ctx.ServeFS(f, p, index, headers)
	switch {
	case errors.Is(err, fs.ErrPermission):
		return Status(http.StatusForbidden)
	case errors.Is(err, fs.ErrNotExist):
		return Status(http.StatusNotFound)
	case err != nil:
		return ctx.Error(http.StatusInternalServerError, err)
	}
	return nil
}

// Now 返回当前时间
//
// 与 time.Now() 的区别在于 Now() 基于当前时区
func (ctx *Context) Now() time.Time { return time.Now().In(ctx.Location) }

// ParseTime 分析基于当前时区的时间
func (ctx *Context) ParseTime(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, ctx.Location)
}
