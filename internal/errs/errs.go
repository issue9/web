// SPDX-License-Identifier: MIT

// Package errs 提供额外的错误处理功能
package errs

import "errors"

type errs []error

// Errors 合并多个非空错误为一个错误
//
// 有关 Is 和 As 均是按顺序找第一个返回 true 的即返回。
func Errors(err ...error) error {
	all := make([]error, 0, len(err))
	for _, e := range err {
		if e != nil {
			all = append(all, e)
		}
	}
	if len(all) == 0 {
		return nil
	}
	return errs(all)
}

func (e errs) Error() string {
	var s string
	for _, err := range e {
		s += err.Error() + "\n"
	}
	return s
}

func (e errs) Is(target error) bool {
	for _, err := range e {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

func (e errs) As(target any) bool {
	for _, err := range e {
		if errors.As(err, &target) {
			return true
		}
	}
	return false
}
