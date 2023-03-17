// SPDX-License-Identifier: MIT

package filter

// SanitizeFuncOf 数据修正的函数原型
//
// 实现方直接修改传递进来的参数即可。
type SanitizeFuncOf[T any] func(*T)

func Sanitizers[T any](s ...func(*T)) func(*T) {
	return func(v *T) {
		for _, ss := range s {
			ss(v)
		}
	}
}
