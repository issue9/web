// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package problems

// Error 与指定的 Problem id 关联的错误对象
type Error struct {
	error
	id ID
}

// NewError 返回 [Error] 错误对象
func NewError(problemID ID, err error) error {
	if err == nil {
		panic("参数 err 不能为空")
	}

	return &Error{id: problemID, error: err}
}

func (err *Error) ID() ID { return err.id }

func (err *Error) Unwrap() error { return err.error }

func (err *Error) Error() string { return err.error.Error() }
