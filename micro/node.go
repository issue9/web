// SPDX-License-Identifier: MIT

package micro

import (
	"time"

	"github.com/issue9/web"
)

// Node 单个服务节点
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

func (n *Node) Serve() error {
	if err := n.Server.Serve(); err != nil {
		return err
	}

	// TODO 注册服务

	return nil
}

func (n *Node) Close(shutdown time.Duration) {
	n.Server.Close(shutdown)
	// TODO <- wait  serve exit

	// TODO 从 registry 退出
}
