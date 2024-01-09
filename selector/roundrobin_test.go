// SPDX-License-Identifier: MIT

package selector

import (
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
)

var _ Selector = &roundRobin{}

func TestRoundRobin(t *testing.T) {
	a := assert.New(t, false)

	sel := NewRoundRobin(false, 0)
	addr, err := sel.Next()
	a.Equal(err, ErrNoPeer()).Empty(addr)

	// 单个节点

	p1 := NewPeer("https://example.com")
	a.NotError(sel.Add(p1))

	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "https://example.com")
	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "https://example.com")

	// 多个节点

	p2 := NewPeer("https://example.io/")
	a.NotError(sel.Add(p2))
	a.Equal(sel.Add(p2), localeutil.Error("has dup peer %s", p2.Addr())) // 添加已存在的节点

	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "https://example.io")
	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "https://example.com")
	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "https://example.io")
	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "https://example.com")

	// 删除节点

	a.NotError(sel.Del(p1.Addr()))

	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "https://example.io")
	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "https://example.io")

	a.NotError(sel.Del(p1.Addr())) // 删除不存在的节点
	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "https://example.io")

	a.NotError(sel.Del(p2.Addr())) // 节点已全部删除
	addr, err = sel.Next()
	a.Equal(err, ErrNoPeer()).Empty(addr)
}
