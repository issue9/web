// SPDX-FileCopyrightText: 2022-2026 caixw
//
// SPDX-License-Identifier: MIT

package validators

import (
	"testing"

	"github.com/issue9/assert/v6"
)

func TestNot(t *testing.T) {
	a := assert.New(t, false)

	a.True(Zero(0)).False(Zero(1))

	nz := Not(Zero[int])
	a.False(nz(0)).True(nz(1))
}

func TestAnd_Or(t *testing.T) {
	a := assert.New(t, false)

	and := And(Between(0, 100), Between(-1, 50))
	a.False(and(0)).
		True(and(1)).
		False(and(51))

	or := Or(Between(0, 100), Between(-1, 50))
	a.True(or(0)).
		True(or(1)).
		False(or(500))
}

func TestZero(t *testing.T) {
	a := assert.New(t, false)

	a.True(Zero(0)).
		True(Zero("")).
		False(Zero(1))
}

func TestNil(t *testing.T) {
	a := assert.New(t, false)
	var pi *int
	var pii *int = nil
	var x any

	a.True(Nil(x)).
		True(Nil(pi)).
		True(Nil(pii)).
		False(Nil(5))
}

func TestRegexp(t *testing.T) {
	a := assert.New(t, false)
	v := Regexp("^\\d+$")

	a.True(v("5")).
		False(v("y5y"))
}
