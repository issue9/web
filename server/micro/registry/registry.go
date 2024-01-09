// SPDX-License-Identifier: MIT

// Package registry 服务注册
package registry

import (
	"net/http/httputil"

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

type Mapper = map[string]web.RouterMatcher
