// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package sanitizers

type RealNumber interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Abs 将数值转换为绝对值
func Abs[T RealNumber](v *T) {
	if *v < 0 {
		*v = -*v
	}
}
