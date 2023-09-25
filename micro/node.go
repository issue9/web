// SPDX-License-Identifier: MIT

package micro

import "github.com/issue9/web"

type Node struct {
	*web.Server
}

type NodeOptions struct {
	// TODO
	Single bool // 不允许多例
}

// NewNode 将 [web.Server] 作为微服务节点
func NewNode(s *web.Server, o *NodeOptions) *Node {
	return &Node{
		Server: s,
	}
}
