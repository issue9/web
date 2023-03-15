// SPDX-License-Identifier: MIT

package validation

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

	z := ValidatorFuncOf[int](zero[int])
	a.True(z.IsValid(0)).False(z.IsValid(1))

	nz := Not[int](z)
	a.False(nz.IsValid(0)).True(nz.IsValid(1))
}

func TestAnd_Or(t *testing.T) {
	a := assert.New(t, false)

	and := And[int](ValidatorFuncOf[int](zero[int]), between(-1, 100))
	a.True(and.IsValid(0)).
		False(and.IsValid(-1)).
		False(and.IsValid(50))

	or := Or[int](ValidatorFuncOf[int](zero[int]), between(-1, 100))
	a.True(or.IsValid(0)).
		True(or.IsValid(-1)).
		True(or.IsValid(50)).
		False(or.IsValid(-2))
}

func TestAnd_OrFunc(t *testing.T) {
	a := assert.New(t, false)

	and := AndFunc(between(0, 100), between(-1, 50))
	a.True(and.IsValid(0)).
		True(and.IsValid(1)).
		False(and.IsValid(51))

	or := OrFunc(between(0, 100), between(-1, 50))
	a.True(or.IsValid(0)).
		True(or.IsValid(1)).
		False(or.IsValid(500))
}
