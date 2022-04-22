// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"fmt"
	"testing"

	"github.com/issue9/assert/v2"
	"golang.org/x/xerrors"
)

var (
	_ fmt.Formatter     = &stackError{}
	_ xerrors.Formatter = &stackError{}
)

func TestStackError(t *testing.T) {
	a := assert.New(t, false)

	err := StackError(nil)
	a.Nil(err)

	err1 := errors.New("abc")
	err = StackError(err1)
	a.ErrorIs(err, err1)
	s := fmt.Sprintf("%v", err)
	a.NotContains(s, "26")
	s = fmt.Sprintf("%+v", err)
	a.Contains(s, "26")

	// 二次包装
	err = StackError(err)
	a.ErrorIs(err, err1)
	s = fmt.Sprintf("%v", err)
	a.NotContains(s, "26")
	s = fmt.Sprintf("%+v", err)
	a.Contains(s, "26")
}
