// SPDX-License-Identifier: MIT

// Package errs 对错误信息的二次处理
package errs

import "fmt"

// Merge 合并多个错误实例
//
// 当两个值都不为 nil 时，会将 origin 作为其底层的错误类型，否则返回其中不为 nil 的值。
func Merge(origin, err error) error {
	if err == nil {
		return origin
	}

	if origin == nil {
		return err
	}

	return fmt.Errorf("在返回 %w 时再次出现错误 %s", origin, err.Error())
}
