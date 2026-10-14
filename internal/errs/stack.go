// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package errs

import (
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
func NewDepthStackError(depth int, err error) error {
	if err == nil {
		return nil
	}

	// 只有 err 本身是 stackError 才直接返回，
	// 如果是链上的其它节点为 stackError，依然还是要对其进行包装。
	if se, ok := err.(*stackError); ok {
		return se
	}

	return &stackError{
		err:   err,
		frame: xerrors.Caller(depth),
	}
}

func (err *stackError) Error() string { return err.err.Error() }

// Format implement fmt.Formatter
func (err *stackError) Format(f fmt.State, c rune) {
	xerrors.FormatError(err, f, c)
}

// FormatError implement xerrors.Formatter
func (err *stackError) FormatError(p xerrors.Printer) error {
	if p.Detail() {
		err.frame.Format(p)
	}
	return err.err
}

func (err *stackError) Unwrap() error { return err.err }
