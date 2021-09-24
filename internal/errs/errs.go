// SPDX-License-Identifier: MIT

// Package errs 对错误信息的二次处理
package errs

import "fmt"

// Merge 合并多个错误实例
func Merge(origin, err error) error {
	if err == nil {
		return origin
	}
	return fmt.Errorf("在返回 %w 时再次出现错误 %s", origin, err.Error())
}
