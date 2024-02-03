// SPDX-License-Identifier: MIT

package selector

import (
	"testing"

	"github.com/issue9/assert/v3"
)

var _ Selector = &random{}

func TestRandom(t *testing.T) {
	a := assert.New(t, false)

	sel := NewRandom(false, 0)
	addr, err := sel.Next()
	a.Equal(err, ErrNoPeer()).Empty(addr)

	// 单个节点

	p1 := NewPeer("https://example.com")
	sel.Update(p1)

	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "https://example.com")
	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "https://example.com")

	// 多个节点

	p2 := NewPeer("https://example.io/")
	sel.Update(p1, p2)

	addr, err = sel.Next()
	a.NotError(err).NotEmpty(addr)
	addr, err = sel.Next()
	a.NotError(err).NotEmpty(addr)
	addr, err = sel.Next()
	a.NotError(err).NotEmpty(addr)
	addr, err = sel.Next()
	a.NotError(err).NotEmpty(addr)

	// 删除节点

	sel.Update()
	addr, err = sel.Next()
	a.Equal(err, ErrNoPeer()).Empty(addr)
	addr, err = sel.Next()
	a.Equal(err, ErrNoPeer()).Empty(addr)
}