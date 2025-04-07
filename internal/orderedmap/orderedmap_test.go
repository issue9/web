// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

package orderedmap

import (
	"maps"
	"testing"

	"github.com/issue9/assert/v4"
)

func TestOrderedMap(t *testing.T) {
	a := assert.New(t, false)

	m := New[int](10)
	a.Equal(m.Len(), 0)

	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)
	a.Equal(m.Len(), 3).
		Equal(maps.Collect(m.Range()), map[string]int{"a": 1, "b": 2, "c": 3})

	v, f := m.Get("b")
	a.True(f).Equal(v, 2)

	m.Delete("b")
	a.Equal(m.Len(), 2)
	v, f = m.Get("b")
	a.False(f).
		Equal(maps.Collect(m.Range()), map[string]int{"a": 1, "c": 3})
}
