// SPDX-License-Identifier: MIT

// Package errs 对错误信息的二次处理
package errs

import "fmt"

// Wrap 包装多个错误信息
func Wrap(origin, new error) error {
	return fmt.Errorf("在返回 %w 时再次出现错误 %s", origin, new.Error())
}
