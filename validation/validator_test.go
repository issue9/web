// SPDX-License-Identifier: MIT

package validation

import (
	"reflect"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/sliceutil"
)

type (
	root1 struct {
		Root *root2
		F1   int
	}

	root2 struct {
		O1 *object
		O2 *object
	}

	object struct {
		Name string
		Age  int
	}
)

func required(v any) bool { return !reflect.ValueOf(v).IsZero() }

func min(v int) ValidateFunc {
	return func(a any) bool {
		if num, ok := a.(int); ok {
			return num >= v
		}
		return false
	}
}

func max(v int) ValidateFunc {
	return func(a any) bool {
		if num, ok := a.(int); ok {
			return num < v
		}
		return false
	}
}

func in(element ...int) ValidateFunc {
	return func(v any) bool {
		return sliceutil.Exists(element, func(elem int) bool { return elem == v })
	}
}

func notIn(element ...int) ValidateFunc {
	return func(v any) bool {
		return !sliceutil.Exists(element, func(elem int) bool { return elem == v })
	}
}

func TestAnd(t *testing.T) {
	a := assert.New(t, false)

	v := And(in(1, 2, 3), notIn(2, 3, 4))
	a.True(v.IsValid(1))
	a.False(v.IsValid(2))
	a.False(v.IsValid(-1))
	a.False(v.IsValid(100))
}

func TestOr(t *testing.T) {
	a := assert.New(t, false)

	v := Or(in(1, 2, 3), notIn(2, 3, 4))
	a.True(v.IsValid(1))
	a.True(v.IsValid(2))
	a.False(v.IsValid(4))
	a.True(v.IsValid(-1))
	a.True(v.IsValid(100))
}

func TestAndF(t *testing.T) {
	a := assert.New(t, false)

	v := AndFunc(in(1, 2, 3).IsValid, notIn(2, 3, 4).IsValid)
	a.True(v.IsValid(1))
	a.False(v.IsValid(2))
	a.False(v.IsValid(-1))
	a.False(v.IsValid(100))
}

func TestOrF(t *testing.T) {
	a := assert.New(t, false)

	v := OrFunc(in(1, 2, 3).IsValid, notIn(2, 3, 4).IsValid)
	a.True(v.IsValid(1))
	a.True(v.IsValid(2))
	a.False(v.IsValid(4))
	a.True(v.IsValid(-1))
	a.True(v.IsValid(100))
}
