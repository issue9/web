// SPDX-License-Identifier: MIT

package selector

import (
	"math/rand"
	"slices"
	"sync"

	"github.com/issue9/localeutil"
)

type weightedRandom struct {
	peers       []string
	weights     []int
	sumOfWeight int
	mux         sync.RWMutex
}

func newWeightedRandom(cap int) Selector {
	return &weightedRandom{
		weights: make([]int, 0, cap),
		peers:   make([]string, 0, cap),
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
		return s.peers[0], nil
	default:
		weight := rand.Intn(s.sumOfWeight) + 1 // 排除 0，TODO(go1.22): 可以采用 rand/v2
		for i, addr := range s.peers {
			weight -= s.weights[i]
			if weight <= 0 {
				return addr, nil
			}
		}
		return s.peers[len(s.peers)-1], nil
	}
}

func (s *weightedRandom) Del(p string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.peers = slices.DeleteFunc(s.peers, func(i string) bool { return i == p })
	return nil
}

func (s *weightedRandom) Add(p Peer) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	wp, ok := p.(WeightedPeer)
	if !ok {
		panic("p 必须实现 WeightedPeer 接口")
	}

	if index := slices.Index(s.peers, wp.Addr()); index >= 0 { // 有重复项
		return localeutil.Error("has dup peer %s", wp.Addr())
	}
	s.weights = append(s.weights, wp.Weight())
	s.peers = append(s.peers, wp.Addr())
	s.sumOfWeight += wp.Weight()
	return nil
}
