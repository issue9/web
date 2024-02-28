// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package selector

import (
	"testing"

	"github.com/issue9/assert/v4"
)

var _ Selector = &roundRobin{}

func TestRoundRobin(t *testing.T) {
	a := assert.New(t, false)

	sel := NewRoundRobin(false, 0)
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
	a.NotError(err).Equal(addr, "https://example.io")
	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "https://example.com")
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
