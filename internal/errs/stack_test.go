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

	err := StackError(nil)
	a.Nil(err)

	err1 := &cerr{"abc"}
	err = StackError(err1)

	a.ErrorIs(err, err1)
	s := fmt.Sprintf("%v", err)
	a.NotContains(s, "33")
	s = fmt.Sprintf("%+v", err)
	a.Contains(s, "33")

	target := &cerr{}
	a.True(errors.As(err, &target)).Equal(target.Error(), err1.Error())

	// 二次包装

	err = StackError(err)

	a.ErrorIs(err, err1)
	s = fmt.Sprintf("%v", err)
	a.NotContains(s, "33")
	s = fmt.Sprintf("%+v", err)
	a.Contains(s, "33")

	target = &cerr{}
	a.True(errors.As(err, &target)).Equal(target.Error(), err1.Error())
}
