// SPDX-License-Identifier: MIT

package selector

import (
	"math/rand"
	"sync"
)

type weightedRandom struct {
	peers       []WeightedPeer
	weights     []int
	sumOfWeight int
	mux         sync.RWMutex
}

func newWeightedRandom(cap int) Updateable {
	return &weightedRandom{
		weights: make([]int, 0, cap),
		peers:   make([]WeightedPeer, 0, cap),
	}
}

func (s *weightedRandom) Next() (string, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	l := len(s.peers)
	switch l {
	case 0:
		return "", ErrNoPeer()
	case 1:
		return s.peers[0].Addr(), nil
	default:
		weight := rand.Intn(s.sumOfWeight) + 1 // 排除 0，TODO(go1.22): 可以采用 rand/v2
		for i, p := range s.peers {
			weight -= s.weights[i]
			if weight <= 0 {
				return p.Addr(), nil
			}
		}
		return s.peers[len(s.peers)-1].Addr(), nil
	}
}

func (s *weightedRandom) Update(peers ...Peer) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.peers = s.peers[:0]
	s.weights = s.weights[:0]
	s.sumOfWeight = 0
	for _, p := range peers {
		wp, ok := p.(WeightedPeer)
		if !ok {
			panic("p 必须实现 WeightedPeer 接口")
		}
		s.sumOfWeight += wp.Weight()
		s.peers = append(s.peers, wp)
		s.weights = append(s.weights, wp.Weight())
	}
}
