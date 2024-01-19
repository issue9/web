// SPDX-License-Identifier: MIT

package selector

import "sync"

// https://github.com/nginx/nginx/commit/52327e0627f49dbda1e8db695e63a4b0af4448b1#diff-3f2250b728a3f5fe1e2d31cbf63c2268R527
type weightedRoundRobin struct {
	currWeight      []int
	effectiveWeight []int
	peers           []WeightedPeer
	mux             sync.RWMutex
}

func newWeightedRoundRobin(cap int) Updateable {
	return &weightedRoundRobin{
		currWeight:      make([]int, 0, cap),
		effectiveWeight: make([]int, 0, cap),
		peers:           make([]WeightedPeer, 0, cap),
	}
}

func (s *weightedRoundRobin) Next() (string, error) {
	var total int
	best := -1 // 最佳节点的索引

	s.mux.RLock()
	defer s.mux.RUnlock()

	switch len(s.peers) {
	case 0:
		return "", ErrNoPeer()
	case 1:
		return s.peers[0].Addr(), nil
	default:
		for i := range s.peers {
			s.currWeight[i] += s.effectiveWeight[i]
			total += s.effectiveWeight[i]

			if s.effectiveWeight[i] < s.currWeight[i] {
				s.effectiveWeight[i] += 1
			}

			if best == -1 || s.currWeight[i] > s.currWeight[best] {
				best = i
			}
		}

		s.currWeight[best] -= total
		return s.peers[best].Addr(), nil
	}
}

func (s *weightedRoundRobin) Update(peers ...Peer) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.peers = s.peers[:0]
	s.effectiveWeight = s.effectiveWeight[:0]
	s.currWeight = s.currWeight[:0]
	for _, p := range peers {
		wp, ok := p.(WeightedPeer)
		if !ok {
			panic("p 必须实现 WeightedPeer 接口")
		}

		s.peers = append(s.peers, wp)
		s.effectiveWeight = append(s.effectiveWeight, wp.Weight())
		s.currWeight = append(s.currWeight, 0)
	}
}
