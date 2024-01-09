// SPDX-License-Identifier: MIT

// Package selector 提供负载均衡的相关功能
package selector

import "github.com/issue9/localeutil"

var errNoPeer = localeutil.Error("no available peer")

type (
	// Selector 负载均衡接口
	Selector interface {
		// Next 返回下一个可用的节点地址
		Next() (string, error)

		// Add 添加新的节点
		//
		// 不允许添加 [Peer.Addr] 返回值相同的项，否则会返回错误。
		Add(Peer) error

		// Del 根据节点地址删除节点
		Del(string) error
	}

	// Peer 后端节点对象
	Peer interface {
		// Addr 返回节点地址
		//
		// 返回值应该是一个有效的地址，且要满足以下条件：
		//  - 不能以 / 结尾；
		// 比如 https://example.com:8080/s1、/path。
		Addr() string
	}

	// WeightedPeer 带权重的节点对象
	WeightedPeer interface {
		Peer

		// Weight 节点的权重值
		Weight() int
	}

	stringPeer string

	weightedPeer struct {
		addr   string
		weight int
	}
)

func (p stringPeer) Addr() string { return string(p) }

func (p weightedPeer) Addr() string { return p.addr }

func (p weightedPeer) Weight() int { return p.weight }

// NewPeer 声明 [Peer] 对象
//
// addr 为节点的地址；
// weight 为该节点的权重，如果未指定该值，则返回普通的 [Peer] 对象，
// 若是指定该值则会返回 [WeightedPeer] 接口。
func NewPeer(addr string, weight ...int) Peer {
	if last := len(addr) - 1; addr[last] == '/' {
		addr = addr[:last]
	}

	switch len(weight) {
	case 0:
		return stringPeer(addr)
	case 1:
		return &weightedPeer{addr: addr, weight: weight[0]}
	default:
		panic("参数 weight 过多")
	}
}

// ErrNoPeer 表示没有有效的节点
func ErrNoPeer() error { return errNoPeer }
