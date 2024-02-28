// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package selector

import (
	"math/rand"
	"sync"
)

type random struct {
	peers []Peer
	mux   sync.RWMutex
}

// NewRandom 返回随机算法的负载均衡实现
//
// weight 是否采用加权算法，如果此值为 true，
// 在调用 [Selector.Add] 时参数必须实现 [WeightedPeer]。
func NewRandom(weight bool, cap int) Updateable {
	if weight {
		return newWeightedRandom(cap)
	}

	return &random{
		peers: make([]Peer, 0, cap),
	}
}

func (s *random) Next() (string, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	l := len(s.peers)
	switch l {
	case 0:
		return "", ErrNoPeer()
	case 1:
		return s.peers[0].Addr(), nil
	default:
		return s.peers[rand.Intn(len(s.peers))].Addr(), nil // TODO(go1.22): 可以采用 rand/v2
	}
}

func (s *random) Update(peers ...Peer) {
	s.mux.Lock()
	s.peers = append(s.peers[:0], peers...)
	s.mux.Unlock()
}
