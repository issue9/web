// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package selector 提供负载均衡的相关功能
package selector

import (
	"bytes"
	"encoding"
	"strconv"

	"github.com/issue9/localeutil"
)

var errNoPeer = localeutil.Error("no available peer")

type (
	// Selector 负载均衡接口
	Selector interface {
		// Next 返回下一个可用的节点地址
		Next() (string, error)
	}

	// Updateable 可动态更新节点信息的负载均衡接口
	Updateable interface {
		Selector

		// Update 更新节点信息
		//
		// 如果传递空数组，则清空节点。
		Update(...Peer)
	}

	// Peer 后端节点对象
	//
	// NOTE: 节点的序列化内容中不能包含半角分号（;）。
	Peer interface {
		encoding.TextMarshaler
		encoding.TextUnmarshaler

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

	stringPeer struct {
		string
	}

	weightedPeer struct {
		addr   string
		weight int
	}
)

func (p *stringPeer) Addr() string { return p.string }

func (p *stringPeer) MarshalText() ([]byte, error) { return []byte(p.string), nil }

func (p *stringPeer) UnmarshalText(data []byte) error {
	p.string = string(data)
	return nil
}

func (p weightedPeer) Addr() string { return p.addr }

func (p weightedPeer) Weight() int { return p.weight }

func (p weightedPeer) MarshalText() ([]byte, error) {
	return []byte(p.addr + "," + strconv.Itoa(p.weight)), nil
}

func (p *weightedPeer) UnmarshalText(data []byte) (err error) {
	index := bytes.LastIndexByte(data, ',')
	if index <= 0 {
		return localeutil.Error("invalid data %s", strconv.Quote(string(data)))
	}
	p.addr = string(data[:index])
	p.weight, err = strconv.Atoi(string(data[index+1:]))
	return err
}

func NewPeer(addr string) Peer {
	if addr == "" {
		return &stringPeer{}
	}

	if last := len(addr) - 1; addr[last] == '/' {
		addr = addr[:last]
	}
	return &stringPeer{addr}
}

func NewWeightedPeer(addr string, weight int) WeightedPeer {
	if addr == "" {
		return &weightedPeer{}
	}

	if last := len(addr) - 1; addr[last] == '/' {
		addr = addr[:last]
	}
	return &weightedPeer{addr: addr, weight: weight}
}

// ErrNoPeer 表示没有有效的节点
func ErrNoPeer() error { return errNoPeer }
