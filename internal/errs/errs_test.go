// SPDX-License-Identifier: MIT

package errs

import (
	"errors"
	"strings"
	"testing"

	"github.com/issue9/assert/v3"
)

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
