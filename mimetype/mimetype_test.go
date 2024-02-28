// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package mimetype

import (
	"errors"
	"testing"

	"github.com/issue9/assert/v4"
)

func TestErrUnsupported(t *testing.T) {
	a := assert.New(t, false)

	a.ErrorIs(ErrUnsupported(), errors.ErrUnsupported)
	a.Equal(ErrUnsupported().Error(), "unsupported serialization")
}
