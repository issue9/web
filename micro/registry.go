// SPDX-License-Identifier: MIT

package micro

// Registry 服务注册与发现需要实现的接口
type Registry interface {
	// Register 注册服务
	Register() error

	// Deregister 注消服务
	Deregister() error

	// Discover 返回指定名称的服务节点
	Discover(name string) []*Node
}
