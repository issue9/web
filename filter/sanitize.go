// SPDX-License-Identifier: MIT

package filter

type SanitizeFuncOf[T any] func(*T)

func Sanitizers[T any](s ...func(*T)) func(*T) {
	return func(v *T) {
		for _, ss := range s {
			ss(v)
		}
	}
}
