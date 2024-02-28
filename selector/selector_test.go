// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package selector

import (
	"testing"

	"github.com/issue9/assert/v4"
)

var (
	_ Peer         = &stringPeer{}
	_ WeightedPeer = &weightedPeer{}
)

func TestNewPeer(t *testing.T) {
	a := assert.New(t, false)
	addr := "http://localhost:8080"

	p := NewPeer(addr)
	sp, ok := p.(*stringPeer)
	a.True(ok).Equal(sp.Addr(), addr)

	p = NewWeightedPeer(addr+"/", 5)
	wp, ok := p.(WeightedPeer)
	a.True(ok).
		Equal(wp.Addr(), addr).
		Equal(wp.Weight(), 5)

	p = NewPeer("")
	sp, ok = p.(*stringPeer)
	a.True(ok).Empty(sp.Addr())

	p = NewWeightedPeer("", 5)
	wp, ok = p.(WeightedPeer)
	a.True(ok).Empty(sp.Addr()).Zero(wp.Weight())
}

func TestStringPeer(t *testing.T) {
	a := assert.New(t, false)
	addr := "http://localhost:8080"
	p := NewPeer(addr)

	data, err := p.MarshalText()
	a.NotError(err).Equal(string(data), addr)

	sp := NewPeer("")
	a.NotError(sp.UnmarshalText(data))
	a.Equal(sp.Addr(), p.Addr())
}

func TestWeightedPeer(t *testing.T) {
	a := assert.New(t, false)
	addr := "http://localhost:8080"
	p := NewWeightedPeer(addr, 5)

	data, err := p.MarshalText()
	a.NotError(err).Equal(string(data), addr+",5")

	pp := &weightedPeer{}
	a.NotError(pp.UnmarshalText(data))
	a.Equal(pp.Addr(), p.Addr()).
		Equal(pp.Weight(), p.Weight())
}
