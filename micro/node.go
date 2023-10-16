// SPDX-License-Identifier: MIT

package micro

import "github.com/issue9/web"

type Node struct {
	web.Server

	nodes []*web.Client // 其它节点
}

type NodeOptions struct {
	// TODO
	Single bool // 不允许多例
}

// NewNode 将 [web.Server] 作为微服务节点
func NewNode(s web.Server, o *NodeOptions) *Node {
	return &Node{
		Server: s,
	}
}
