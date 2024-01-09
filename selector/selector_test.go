// SPDX-License-Identifier: MIT

package selector

import (
	"testing"

	"github.com/issue9/assert/v3"
)

var _ Peer = stringPeer("localhost:8080")

func TestNewPeer(t *testing.T) {
	a := assert.New(t, false)
	addr := "http://localhost:8080"

	p := NewPeer(addr)
	sp, ok := p.(stringPeer)
	a.True(ok).Equal(sp.Addr(), addr)

	p = NewPeer(addr+"/", 5)
	wp, ok := p.(WeightedPeer)
	a.True(ok).
		Equal(wp.Addr(), addr).
		Equal(wp.Weight(), 5)

	a.PanicString(func() {
		NewPeer(addr, 1, 2, 3)
	}, "参数 weight 过多")
}
