// SPDX-License-Identifier: MIT

// Package registry 服务注册
package registry

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/issue9/web"
	"github.com/issue9/web/selector"
)

// Registry 服务注册与发现需要实现的接口
type Registry interface {
	// Register 注册服务
	//
	// name 服务名称；
	// peer 节点信息；
	//
	// 返回一个用于注销当前服务的方法；
	Register(name string, peer selector.Peer) (DeregisterFunc, error)

	// Discover 返回指定名称的服务节点
	Discover(name string) (selector.Selector, error)

	// ReverseProxy 返回反向代理对象
	ReverseProxy(Mapper) *httputil.ReverseProxy
}

type DeregisterFunc = func() error

type Mapper map[string]web.RouterMatcher

// 将 Mapper 转换为 httputil.ProxyRequest.Rewrite 字段类型的函数
//
// f 为查找指定名称的 [selector.Selector] 对象。当 Mapper 中找到匹配的项时，
// 需要通过 f 找对应的 [selector.Selector] 对象。
func (ms Mapper) asRewrite(f func(string) selector.Selector) func(r *httputil.ProxyRequest) {
	return func(r *httputil.ProxyRequest) {
		for name, match := range ms {
			if !match.Match(r.In, nil) {
				continue
			}

			s := f(name)
			if s == nil {
				panic(web.NewError(http.StatusNotFound, web.NewLocaleError("not found micro service %s", name)))
			}

			route, err := s.Next()
			if err != nil {
				panic(err)
			}

			u, err := url.Parse(route)
			if err != nil {
				panic(err) // Selector 实现得不标准
			}

			r.SetURL(u)
			// r.Out.Host = r.In.Host

			break
		}
	}
}
