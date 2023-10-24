// SPDX-License-Identifier: MIT

package micro

import "github.com/issue9/web"

// Registry 服务注册与发现需要实现的接口
type Registry interface {
	// Register 注册服务
	Register(Options) error

	// Deregister 注消服务
	//
	// name 表示注册的服务名称；
	// id 表当前注册服务实例的唯一 ID；
	Deregister(Options) error

	// Discover 返回指定名称的服务节点
	Discover(name string) (web.Selector, error)

	Services() ([]Options, error)
}

type Options struct {
	ID      string            // 表当前注册服务实例的唯一 ID
	Name    string            // 表示注册的服务名称
	URL     string            // 注册服务的地址
	Matcher web.RouterMatcher // 作为外放接口的匹配接口
}
