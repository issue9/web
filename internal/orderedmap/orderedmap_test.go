// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

package orderedmap

import (
	"encoding/json"
	"maps"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/issue9/assert/v4"
)

var (
	_ json.Marshaler      = &OrderedMap[int]{}
	_ yaml.BytesMarshaler = &OrderedMap[int]{}
)

func TestOrderedMap(t *testing.T) {
	a := assert.New(t, false)

	m := New[int](10)
	a.Equal(m.Len(), 0)

	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)
	a.Equal(m.Len(), 3).
		Equal(maps.Collect(m.Iter()), map[string]int{"a": 1, "b": 2, "c": 3})

	v, f := m.Get("b")
	a.True(f).Equal(v, 2)

	m.Delete("b")
	a.Equal(m.Len(), 2)
	_, f = m.Get("b")
	a.False(f).
		Equal(maps.Collect(m.Iter()), map[string]int{"a": 1, "c": 3})
}

func TestOrderedMap_Marshal(t *testing.T) {
	a := assert.New(t, false)

	m := New[int](10)
	a.Equal(m.Len(), 0)

	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)

	t.Run("JSON", func(t *testing.T) {
		data, err := m.MarshalJSON()
		a.NotError(err).Equal(string(data), `{"a":1,"b":2,"c":3}`)
	})

	t.Run("YAML", func(t *testing.T) {
		data, err := m.MarshalYAML()
		a.NotError(err).Equal(string(data), `a:1
b:2
c:3
`)
	})
}
