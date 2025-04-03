// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package testdata

// object desc
type object struct {
	Field1 int    // f1
	field2 string // f2

	Object *obj2 `yaml:"object"` // obj
}

// obj2 desc
type obj2 struct {
	Field int
}
