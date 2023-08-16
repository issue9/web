// SPDX-License-Identifier: MIT

package errs

import (
	"errors"
	"fmt"
	"testing"

	"github.com/issue9/assert/v3"
	"golang.org/x/xerrors"
)

var (
	_ fmt.Formatter     = &stackError{}
	_ xerrors.Formatter = &stackError{}
)

type cerr struct {
	msg string
}

func (err *cerr) Error() string { return err.msg }

func TestStackError(t *testing.T) {
	a := assert.New(t, false)

	err := NewStackError(nil)
	a.Nil(err)

	err1 := &cerr{"abc"}
	err = NewStackError(err1)

	a.ErrorIs(err, err1)
	s := fmt.Sprintf("%v", err)
	a.NotContains(s, "32") // 依赖调用 NewStackError 的行号
	s = fmt.Sprintf("%+v", err)
	a.Contains(s, "32") // 依赖调用 NewStackError 的行号

	var target1 *cerr
	a.True(errors.As(err, &target1)).Equal(target1.Error(), err1.Error())

	// 二次包装

	err = NewStackError(err)

	a.ErrorIs(err, err1)
	s = fmt.Sprintf("%v", err)
	a.NotContains(s, "32") // 依赖调用 NewStackError 的行号
	s = fmt.Sprintf("%+v", err)
	a.Contains(s, "32") // 依赖调用 NewStackError 的行号

	var target2 *cerr
	a.True(errors.As(err, &target2)).Equal(target2.Error(), err1.Error())
}
