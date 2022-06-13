// SPDX-License-Identifier: MIT

package web

import (
	"errors"
	"fmt"
	"strings"
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
	a.NotContains(s, "27")
	s = fmt.Sprintf("%+v", err)
	a.Contains(s, "27")

	target := errors.New("")
	a.True(errors.As(err, &target)).Equal(target.Error(), err1.Error())

	// 二次包装

	err = StackError(err)

	a.ErrorIs(err, err1)
	s = fmt.Sprintf("%v", err)
	a.NotContains(s, "27")
	s = fmt.Sprintf("%+v", err)
	a.Contains(s, "27")

	target = errors.New("")
	a.True(errors.As(err, &target)).Equal(target.Error(), err1.Error())
}

func TestErrors(t *testing.T) {
	a := assert.New(t, false)
	err1 := errors.New("err1")
	err2 := StackError(errors.New("err2"))
	err3 := errors.New("err3")
	var err4 error
	all := Errors(err1, err2, err3, err4, nil, StackError(nil))
	a.Equal(3, len(all.(errs)))

	a.True(errors.Is(all, err1))
	a.True(errors.Is(all, err2))
	a.True(errors.Is(all, err3))

	target2 := &stackError{}
	a.True(errors.As(all, &target2))

	s := all.Error()
	a.Equal(3, strings.Count(s, "\n"))
}
