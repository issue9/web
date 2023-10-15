// SPDX-License-Identifier: MIT

package errs

import (
	"errors"
	"fmt"

	"golang.org/x/xerrors"
)

type stackError struct {
	err   error
	frame xerrors.Frame
}

// NewDepthStackError 为 err 带上调用信息
//
// 如果 err 为 nil，则返回 nil。
// 多次调用 NewDepthStackError 包装，则返回第一次包装的返回值。
// depth 为 1 表示调用 NewDepthStackError 的位置。
//
// 如果需要输出调用堆栈信息，需要指定 %+v 标记。
func NewDepthStackError(depth int, err error) error {
	if err == nil {
		return nil
	}

	var se *stackError
	if errors.As(err, &se) {
		return se
	}

	return &stackError{
		err:   err,
		frame: xerrors.Caller(depth),
	}
}

func (err *stackError) Error() string { return err.err.Error() }

// fmt.Formatter
func (err *stackError) Format(f fmt.State, c rune) {
	xerrors.FormatError(err, f, c)
}

// xerrors.Formatter
func (err *stackError) FormatError(p xerrors.Printer) error {
	if p.Detail() {
		err.frame.Format(p)
	}
	return err.err
}

func (err *stackError) Unwrap() error { return err.err }
