// SPDX-FileCopyrightText: 2022-2026 caixw
//
// SPDX-License-Identifier: MIT

package validators

import (
	"testing"

	"github.com/issue9/assert/v6"
)

type object struct {
	Name string
	Age  int
}

func TestIn(t *testing.T) {
	a := assert.New(t, false)

	rule1 := In(1, 2)
	a.False(rule1(3))
	a.True(rule1(1))

	v1 := In(object{}, object{Name: "name"})
	a.True(v1(object{})).
		True(v1(object{Name: "name"})).
		False(v1(object{Name: "name", Age: 1}))

	v2 := In(&object{}, &object{Name: "name"})
	a.False(v2(&object{})).
		False(v2(&object{Name: "name"})).
		False(v2(&object{Name: "name", Age: 1}))
}

func TestNotIn(t *testing.T) {
	a := assert.New(t, false)

	v1 := NotIn(1, 2)
	a.True(v1(3)).
		False(v1(1))

	v2 := NotIn(object{}, object{Name: "name"})
	a.False(v2(object{})).
		False(v2(object{Name: "name"})).
		True(v2(object{Name: "name", Age: 1}))

	v3 := NotIn(&object{}, &object{Name: "name"})
	a.True(v3(&object{})).
		True(v3(&object{Name: "name"})).
		True(v3(&object{Name: "name", Age: 1}))
}
