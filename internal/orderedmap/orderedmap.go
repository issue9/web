// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

// Package orderedmap 提供键名类型为 string 的有序 map
package orderedmap

import (
	"iter"
	"slices"
)

// OrderedMap 键名类型为 string 的有序 map
type OrderedMap[V any] struct {
	items   map[string]V
	ordered []string
}

func New[V any](c int) *OrderedMap[V] {
	return &OrderedMap[V]{
		items:   make(map[string]V, c),
		ordered: make([]string, 0, c),
	}
}

// Set 设置键值对，如果已经存在，则覆盖。
func (m *OrderedMap[V]) Set(key string, value V) {
	if _, ok := m.items[key]; !ok {
		m.ordered = append(m.ordered, key)
	}
	m.items[key] = value
}

// Add 添加新项，如果已经存在，则 panic
func (m *OrderedMap[V]) Add(key string, value V) {
	if _, ok := m.items[key]; ok {
		panic("已经存在同名的键名")
	}

	m.Set(key, value)
}

func (m *OrderedMap[V]) Get(key string) (V, bool) {
	value, ok := m.items[key]
	return value, ok
}

func (m *OrderedMap[V]) Delete(key string) {
	if _, ok := m.items[key]; ok {
		delete(m.items, key)
		m.ordered = slices.DeleteFunc(m.ordered, func(e string) bool { return e == key })
	}
}

func (m *OrderedMap[V]) Range() iter.Seq2[string, V] {
	return func(yield func(string, V) bool) {
		for _, key := range m.ordered {
			if !yield(key, m.items[key]) {
				break
			}
		}
	}
}

func (m *OrderedMap[V]) Len() int { return len(m.items) }
