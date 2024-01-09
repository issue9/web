// SPDX-License-Identifier: MIT

package selector

import (
	"slices"
	"sync"

	"github.com/issue9/localeutil"
)

// https://github.com/nginx/nginx/commit/52327e0627f49dbda1e8db695e63a4b0af4448b1#diff-3f2250b728a3f5fe1e2d31cbf63c2268R527
type weightedRoundRobin struct {
	currWeight      []int
	effectiveWeight []int
	peers           []string
	mux             sync.RWMutex
}

func newWeightedRoundRobin(cap int) Selector {
	return &weightedRoundRobin{
		currWeight:      make([]int, 0, cap),
		effectiveWeight: make([]int, 0, cap),
		peers:           make([]string, 0, cap),
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
		return s.peers[0], nil
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
		return s.peers[best], nil
	}
}

func (s *weightedRoundRobin) Add(p Peer) error {
	wp, ok := p.(WeightedPeer)
	if !ok {
		panic("p 必须实现 WeightedPeer 接口")
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	if index := slices.Index(s.peers, p.Addr()); index >= 0 { // 有重复项
		return localeutil.Error("has dup peer %s", p.Addr())
	}

	s.peers = append(s.peers, wp.Addr())
	s.effectiveWeight = append(s.effectiveWeight, wp.Weight())
	s.currWeight = append(s.currWeight, 0)

	return nil
}

func (s *weightedRoundRobin) Del(addr string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	i := slices.Index(s.peers, addr)
	if i < 0 {
		return nil
	}

	j := i + 1
	s.peers = slices.Delete(s.peers, i, j)
	s.currWeight = slices.Delete(s.currWeight, i, j)
	s.effectiveWeight = slices.Delete(s.effectiveWeight, i, j)

	return nil
}
