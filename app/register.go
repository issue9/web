// SPDX-License-Identifier: MIT

package app

import "slices"

type register[T any] struct {
	items map[string]T
}

func newRegister[T any]() *register[T] {
	return &register[T]{items: make(map[string]T, 5)}
}

// 同名会覆盖
func (r *register[T]) register(v T, name ...string) {
	if len(name) == 0 {
		panic("必须指定至少一个 name 参数")
	}

	if i := slices.Index(name, ""); i >= 0 {
		panic("参数 name 中不能包含空字符串")
	}

	for _, n := range name {
		r.items[n] = v
	}
}

func (r *register[T]) get(name string) (T, bool) {
	v, found := r.items[name]
	return v, found
}
