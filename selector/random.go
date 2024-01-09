// SPDX-License-Identifier: MIT

package selector

import (
	"math/rand"
	"slices"
	"sync"

	"github.com/issue9/localeutil"
	"github.com/issue9/sliceutil"
)

type random struct {
	peers []string
	mux   sync.RWMutex
}

// NewRandom 返回随机算法的负载均衡实现
//
// weight 是否采用加权算法，如果此值为 true，
// 在调用 [Selector.Add] 时参数必须实现 [WeightedPeer]。
func NewRandom(weight bool, cap int) Selector {
	if weight {
		return newWeightedRandom(cap)
	}

	return &random{
		peers: make([]string, 0, cap),
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
		return s.peers[0], nil
	default:
		return s.peers[rand.Intn(len(s.peers))], nil // TODO(go1.22): 可以采用 rand/v2
	}
}

func (s *random) Del(p string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.peers = sliceutil.Delete(s.peers, func(i string, _ int) bool { return i == p })
	return nil
}

func (s *random) Add(p Peer) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if index := slices.Index(s.peers, p.Addr()); index >= 0 { // 有重复项
		return localeutil.Error("has dup peer %s", p.Addr())
	}
	s.peers = append(s.peers, p.Addr())
	return nil
}
