// SPDX-License-Identifier: MIT

package selector

import "sync"

type roundRobin struct {
	curr  int
	peers []Peer
	mux   sync.RWMutex
}

// NewRoundRobin 轮询法
//
// weight 是否采用加权算法，如果此值为 true，
// 在调用 [Selector.Add] 时参数必须实现 [WeightedPeer]。
func NewRoundRobin(weight bool, cap int) Updateable {
	if weight {
		return newWeightedRoundRobin(cap)
	}

	return &roundRobin{
		peers: make([]Peer, 0, cap),
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
		return s.peers[0].Addr(), nil
	default:
		if s.curr++; s.curr >= l {
			s.curr = 0
		}
		return s.peers[s.curr].Addr(), nil
	}
}

func (s *roundRobin) Update(peers ...Peer) {
	s.mux.Lock()

	s.peers = append(s.peers[:0], peers...)
	if s.curr >= len(s.peers) { // 判断 curr 是否超了
		s.curr = 0
	}

	s.mux.Unlock()
}
