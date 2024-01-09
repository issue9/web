// SPDX-License-Identifier: MIT

package selector

import (
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/localeutil"
)

func TestWeightRoundRobin(t *testing.T) {
	a := assert.New(t, false)

	sel := NewRoundRobin(true, 0)
	addr, err := sel.Next()
	a.Equal(err, ErrNoPeer()).Empty(addr)

	a.PanicString(func() {
		sel.Add(NewPeer("https://example.com"))
	}, "p 必须实现 WeightedPeer 接口")

	// 单个节点

	p1 := NewPeer("https://example.com", 1)
	a.NotError(sel.Add(p1))

	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "https://example.com")
	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "https://example.com")

	// 多个节点

	p2 := NewPeer("https://example.io/", 3)
	a.NotError(sel.Add(p2))
	a.Equal(sel.Add(p2), localeutil.Error("has dup peer %s", p2.Addr())) // 添加已存在的节点

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
