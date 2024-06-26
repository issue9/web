// SPDX-FileCopyrightText: 2018-2024 caixw
//
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
	F3 []*S  // 引用自身
}

// S2 Alias
type S2 = S

// G Doc
type G[T any] struct {
	F1 T   // F1 Doc
	F2 int // F2 Doc
	F3 *G[T]

	F4 *GS[int, int, int] // 与 GS 循环引用
}

type GInt G[int]

// GS Doc
type GS[T0 any, T1 any, T2 any] struct {
	G[T1] // 应该被忽略的注释
	F3    T0
	F4    T2
	F5    S2          // 引用类型的字段
	f6    interface{} // 不可导出的接口，应该作为 NotFound 被忽略

	F7 *G[int]
}

// GSNumber Doc
type GSNumber GS[int, string, Int]
