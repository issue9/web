// SPDX-License-Identifier: MIT

package selector

import (
	"slices"
	"sync"

	"github.com/issue9/localeutil"
)

type roundRobin struct {
	curr  int
	peers []string
	mux   sync.RWMutex
}

// NewRoundRobin 轮询法
//
// weight 是否采用加权算法，如果此值为 true，
// 在调用 [Selector.Add] 时参数必须实现 [WeightedPeer]。
func NewRoundRobin(weight bool, cap int) Selector {
	if weight {
		return newWeightedRoundRobin(cap)
	}

	return &roundRobin{
		peers: make([]string, 0, cap),
	}
}

func (s *roundRobin) Next() (string, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	l := len(s.peers)
	switch l {
	case 0:
		return "", ErrNoPeer()
	case 1:
		return s.peers[0], nil
	default:
		if s.curr++; s.curr >= l {
			s.curr = 0
		}
		return s.peers[s.curr], nil
	}
}

func (s *roundRobin) Del(addr string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.peers = slices.DeleteFunc(s.peers, func(i string) bool { return i == addr })

	if s.curr >= len(s.peers) { // 判断 curr 是否超了
		s.curr = 0
	}
	return nil
}

func (s *roundRobin) Add(p Peer) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if index := slices.Index(s.peers, p.Addr()); index >= 0 { // 有重复项
		return localeutil.Error("has dup peer %s", p.Addr())
	}
	s.peers = append(s.peers, p.Addr())
	return nil
}
