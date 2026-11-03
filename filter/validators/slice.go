// SPDX-FileCopyrightText: 2022-2026 caixw
//
// SPDX-License-Identifier: MIT

package validators

import "slices"

// In 创建判断值是否在 element 之中的验证器
func In[T comparable](element ...T) Validator[T] {
	return func(v T) bool { return slices.Index(element, v) >= 0 }
}

// NotIn 创建判断值是否不在 element 之中的验证器
func NotIn[T comparable](element ...T) Validator[T] { return Not(In(element...)) }
