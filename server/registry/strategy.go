// SPDX-License-Identifier: MIT

package registry

import "github.com/issue9/web/selector"

// Strategy 为初始化 [Registry] 对象提供与 [selector.Selector] 相关的方案
type Strategy struct {
	// NewSelector 构建 [selector.Updateable]
	NewSelector func() selector.Updateable

	// NewPeer 构建与 [StrategyNewSelector] 相匹配的 [selector.Peer] 零值对象
	NewPeer func() selector.Peer
}

func NewStrategy(sel func() selector.Updateable, p func() selector.Peer) *Strategy {
	return &Strategy{NewSelector: sel, NewPeer: p}
}

func NewRandomStrategy() *Strategy {
	return NewStrategy(func() selector.Updateable {
		return selector.NewRandom(false, 10)
	}, func() selector.Peer {
		return selector.NewPeer("")
	})
}

func NewWeightedRandomStrategy() *Strategy {
	return NewStrategy(func() selector.Updateable {
		return selector.NewRandom(true, 10)
	}, func() selector.Peer {
		return selector.NewWeightedPeer("", 0)
	})
}

func NewRoundRobinStrategy() *Strategy {
	return NewStrategy(func() selector.Updateable {
		return selector.NewRoundRobin(false, 10)
	}, func() selector.Peer {
		return selector.NewPeer("")
	})
}

func NewWeightedRoundRobinStrategy() *Strategy {
	return NewStrategy(func() selector.Updateable {
		return selector.NewRoundRobin(true, 10)
	}, func() selector.Peer {
		return selector.NewWeightedPeer("", 0)
	})
}
