// SPDX-License-Identifier: MIT

package base

import (
	"fmt"

	"golang.org/x/xerrors"
)

type stackError struct {
	err   error
	frame xerrors.Frame
}

// StackError 为 err 带上调用信息
//
// 位置从调用 StackError 开始。
// 如果 err 为 nil，则返回 nil，如果 err 本身就为 StackError 返回的类型，则原样返回。
//
// 如果需要输出调用堆栈信息，需要指定 %+v 标记。
func StackError(err error) error {
	if err == nil {
		return nil
	}

	if _, ok := err.(*stackError); ok {
		return err
	}

	return &stackError{
		err:   err,
		frame: xerrors.Caller(1),
	}
}

func (err *stackError) Error() string { return err.err.Error() }

func (err *stackError) Format(f fmt.State, c rune) {
	xerrors.FormatError(err, f, c)
}

func (err *stackError) FormatError(p xerrors.Printer) error {
	p.Print(err.err)
	if p.Detail() {
		err.frame.Format(p)
	}
	return nil
}

func (err *stackError) Is(target error) bool { return err.err == target }

func (err *stackError) Unwrap() error { return err.err }
