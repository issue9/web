// SPDX-License-Identifier: MIT

package context

import (
	"net/http"

	"github.com/issue9/mux/v2"
)

// HandlerFunc 路由项处理函数原型
type HandlerFunc func(*Context)

// Prefix 管理带有统一前缀的路由项
type Prefix struct {
	prefix string
	b      *Builder
}

// Router 返回操作路由项的实例
//
// Router 可以处理兼容标准库的 net/http.Handler。
func (b *Builder) Router() *mux.Prefix {
	return b.router
}

// Handle 添加路由请求项
func (b *Builder) Handle(path string, h HandlerFunc, method ...string) error {
	return b.router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		h(b.New(w, r))
	}, method...)
}

// Handle 添加路由请求项
func (b *Builder) handle(path string, h HandlerFunc, method ...string) *Builder {
	if err := b.Handle(path, h, method...); err != nil { // 路由项语法错误，直接 panic
		panic(err)
	}
	return b
}

// Get 添加 GET 请求处理项
func (b *Builder) Get(path string, h HandlerFunc) *Builder {
	return b.handle(path, h, http.MethodGet)
}

// Post 添加 POST 请求处理项
func (b *Builder) Post(path string, h HandlerFunc) *Builder {
	return b.handle(path, h, http.MethodPost)
}

// Put 添加 PUT 请求处理项
func (b *Builder) Put(path string, h HandlerFunc) *Builder {
	return b.handle(path, h, http.MethodPut)
}

// Delete 添加 DELETE 请求处理项
func (b *Builder) Delete(path string, h HandlerFunc) *Builder {
	return b.handle(path, h, http.MethodDelete)
}

// Options 添加 OPTIONS 请求处理项
func (b *Builder) Options(path string, h HandlerFunc) *Builder {
	return b.handle(path, h, http.MethodOptions)
}

// Patch 添加 PATCH 请求处理项
func (b *Builder) Patch(path string, h HandlerFunc) *Builder {
	return b.handle(path, h, http.MethodPatch)
}

func (b *Builder) Prefix(prefix string) *Prefix {
	return &Prefix{
		prefix: prefix,
		b:      b,
	}
}

// Handle 添加路由请求项
func (p *Prefix) Handle(path string, h HandlerFunc, method ...string) error {
	return p.b.Handle(p.prefix+path, h, method...)
}

// Handle 添加路由请求项
func (p *Prefix) handle(path string, h HandlerFunc, method ...string) *Prefix {
	if err := p.Handle(path, h, method...); err != nil {
		panic(err)
	}
	return p
}

// Get 添加 GET 请求处理项
func (p *Prefix) Get(path string, h HandlerFunc) *Prefix {
	return p.handle(path, h, http.MethodGet)
}

// Post 添加 POST 请求处理项
func (p *Prefix) Post(path string, h HandlerFunc) *Prefix {
	return p.handle(path, h, http.MethodPost)
}

// Put 添加 PUT 请求处理项
func (p *Prefix) Put(path string, h HandlerFunc) *Prefix {
	return p.handle(path, h, http.MethodPut)
}

// Delete 添加 DELETE 请求处理项
func (p *Prefix) Delete(path string, h HandlerFunc) *Prefix {
	return p.handle(path, h, http.MethodDelete)
}

// Options 添加 OPTIONS 请求处理项
func (p *Prefix) Options(path string, h HandlerFunc) *Prefix {
	return p.handle(path, h, http.MethodOptions)
}

// Patch 添加 PATCH 请求处理项
func (p *Prefix) Patch(path string, h HandlerFunc) *Prefix {
	return p.handle(path, h, http.MethodPatch)
}
