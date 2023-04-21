// SPDX-License-Identifier: MIT

package filter

import (
	"reflect"
	"testing"

	"github.com/issue9/assert/v3"
)

func zero[T any](v T) bool { return reflect.ValueOf(v).IsZero() }

func between[T ~int | ~uint | float32 | float64](min, max T) ValidatorFuncOf[T] {
	return ValidatorFuncOf[T](func(vv T) bool { return vv >= min && vv <= max })
}

func TestNot(t *testing.T) {
	a := assert.New(t, false)

	z := zero[int]
	a.True(z(0)).False(z(1))

	nz := Not(zero[int])
	a.False(nz(0)).True(nz(1))
}

func TestAnd_OrFunc(t *testing.T) {
	a := assert.New(t, false)

	and := And(between(0, 100), between(-1, 50))
	a.True(and(0)).
		True(and(1)).
		False(and(51))

	or := Or(between(0, 100), between(-1, 50))
	a.True(or(0)).
		True(or(1)).
		False(or(500))
}
