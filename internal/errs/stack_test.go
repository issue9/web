// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package errs

import (
	"errors"
	"fmt"
	"testing"

	"github.com/issue9/assert/v4"
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

func TestNewDepthStackError(t *testing.T) {
	a := assert.New(t, false)

	err := NewDepthStackError(1, nil)
	a.Nil(err)

	err1 := &cerr{"abc"}
	err = NewDepthStackError(1, err1)

	a.ErrorIs(err, err1)
	a.NotContains(fmt.Sprintf("%v", err), "34"). // 依赖调用 NewStackError 的行号
				Contains(fmt.Sprintf("%+v", err), "34"). // 依赖调用 NewStackError 的行号
				Equal(err.Error(), err1.Error())

	var target1 *cerr
	a.True(errors.As(err, &target1)).Equal(target1.Error(), err1.Error())

	// 二次包装

	err = NewDepthStackError(1, err)

	a.ErrorIs(err, err1)
	a.NotContains(fmt.Sprintf("%v", err), "34"). // 依赖调用 NewStackError 的行号
				Contains(fmt.Sprintf("%+v", err), "34") // 依赖调用 NewStackError 的行号

	var target2 *cerr
	a.True(errors.As(err, &target2)).Equal(target2.Error(), err1.Error())
}
