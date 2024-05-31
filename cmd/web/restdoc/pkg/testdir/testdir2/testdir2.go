// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package testdir2

import "github.com/issue9/web/restdoc/pkg"

type String string

type (
	S2 = pkg.S
	S3 = pkg.S2
	// S4
	S4 = S2
	S5 = S3 // S5
	S6 = S4
	S7 = S5

	P1 = *S2 // P1 Doc
	P2 *S3
	P3 *pkg.S2
	P4 = *pkg.S

	// A1 Doc
	A1 = []S2
	A2 []S3
	A3 = []pkg.S2
	A4 []pkg.S
	A5 = [3]S2
)

// GS Doc
type GS[T0 any, T1 any, T2 any] struct {
	pkg.G[T0] // pkg.GS
	F3        T0
	F4        T2
	F5        S2 // 引用类型的字段
	F6        pkg.Int
}

type GSNumber = GS[int, int, pkg.S]
