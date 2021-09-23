// SPDX-License-Identifier: MIT

package errs

import (
	"errors"
	"testing"

	"github.com/issue9/assert"
)

func TestWraf(t *testing.T) {
	a := assert.New(t)

	err1 := errors.New("err1")
	err2 := errors.New("er2")
	err3 := Wrap(err1, err2)
	a.ErrorIs(err3, err1).
		Contains(err3.Error(), err1.Error()).
		Contains(err3.Error(), err2.Error())
}
