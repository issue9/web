// SPDX-License-Identifier: MIT

package testdir

type t1 int

type t2 t1

type (
	t3 uint8
	t4 t3
	t5 struct{}
)

const (
	t1V1 t1 = iota
	t1V2

	t1V3

	other = "5"

	t1V4 t1 = 4
)

const t3V1 t3 = 1

const t1V5 t1 = 5

const V2t3 t3 = 2

const (
	T2v1 t2 = iota
	T2v2
)
