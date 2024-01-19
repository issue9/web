// SPDX-License-Identifier: MIT

package selector

import (
	"testing"

	"github.com/issue9/assert/v3"
)

func TestWeightRoundRobin(t *testing.T) {
	a := assert.New(t, false)

	sel := NewRoundRobin(true, 0)
	addr, err := sel.Next()
	a.Equal(err, ErrNoPeer()).Empty(addr)

	a.PanicString(func() {
		sel.Update(NewPeer("https://example.com"))
	}, "p 必须实现 WeightedPeer 接口")

	// 单个节点

	p1 := NewWeightedPeer("https://example.com", 1)
	sel.Update(p1)

	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "https://example.com")
	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "https://example.com")

	// 多个节点

	p2 := NewWeightedPeer("https://example.io/", 3)
	sel.Update(p1, p2)

	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "https://example.io")
	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "https://example.com")

	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "https://example.io")
	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "https://example.io")
	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "https://example.com")

	// 删除节点

	sel.Update()
	addr, err = sel.Next()
	a.Equal(err, ErrNoPeer()).Empty(addr)
	addr, err = sel.Next()
	a.Equal(err, ErrNoPeer()).Empty(addr)
}
