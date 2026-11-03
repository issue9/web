// SPDX-FileCopyrightText: 2022-2026 caixw
//
// SPDX-License-Identifier: MIT

package validators

import (
	"math"
	"testing"

	"github.com/issue9/assert/v6"
)

func TestBetween(t *testing.T) {
	a := assert.New(t, false)

	a.Panic(func() { Between(100, 5) })

	a.True(BetweenEqual(5, math.MaxInt16)(5)).
		True(BetweenEqual(5.0, 100.0)(5.1)).
		False(BetweenEqual(5, 100)(200)).
		False(BetweenEqual(5, 100)(-1)).
		False(BetweenEqual(5.0, 100.0)(-1.1))

	a.False(Between(5, math.MaxInt16)(5)).
		True(Between(5.0, 100.0)(5.1)).
		False(Between(5, 100)(200)).
		False(Between(5, 100)(-1)).
		False(Between(5.0, 100.0)(100))
}

func TestGreat(t *testing.T) {
	a := assert.New(t, false)

	r := GreatEqual(6)
	a.True(r(6)).
		True(r(10)).
		False(r(5))

	r = Great(6)
	a.False(r(6)).
		True(r(10)).
		False(r(5))
}

func TestLess(t *testing.T) {
	a := assert.New(t, false)

	r := LessEqual(6)
	a.True(r(6)).
		True(r(5)).
		False(r(10))

	r = Less(6)
	a.False(r(6)).
		False(r(10)).
		True(r(5))
}

func TestZeroOr(t *testing.T) {
	a := assert.New(t, false)

	z := ZeroOr(Great(5))
	a.True(z(0)).
		True(z(6)).
		True(z(100)).
		False(z(5)).
		False(z(1)).
		False(z(-1))
}
