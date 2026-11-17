// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package problems

import "unique"

// Error 与指定的 Problem id 关联的错误对象
type Error struct {
	error
	id unique.Handle[string]
}

// NewError 返回 [Error] 错误对象
func NewError(problemID unique.Handle[string], err error) error {
	if err == nil {
		panic("参数 err 不能为空")
	}

	return &Error{id: problemID, error: err}
}

func (err *Error) ID() unique.Handle[string] { return err.id }

func (err *Error) Unwrap() error { return err.error }

func (err *Error) Error() string { return err.error.Error() }
