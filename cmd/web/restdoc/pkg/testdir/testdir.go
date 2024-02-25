// SPDX-License-Identifier: MIT

// Package pkg 测试数据
package pkg

import "time"

type (
	Int int // INT

	uint32 = uint8

	// X
	X = uint32 // XX
)

// S Doc
type S struct { // 此行不会作为 S 的 comment
	Int
	F1 int // INT
	S  struct {
		F2 string
		T  time.Time
	}
	F2 []Int // F2 Doc
	// F3 *S
}

// S2 Alias
type S2 = S

// G Doc
type G[T any] struct {
	F1 T   // F1 Doc
	F2 int // F2 Doc
}

type GInt G[int]

// GS Doc
type GS[T0 any, T1 any, T2 any] struct {
	G[T1] // 应该被忽略的注释
	F3    T0
	F4    T2
	F5    S2 // 引用类型的字段
}

// GSNumber Doc
type GSNumber GS[int, string, Int]
