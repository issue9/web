// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package install

import (
	"errors"
	"testing"

	"github.com/issue9/assert"
)

func TestReturnError(t *testing.T) {
	a := assert.New(t)

	err := errors.New("msg")
	r := ReturnError(err)
	a.Equal(r.message, "msg").
		Equal(r.typ, typeFailed)

	err = nil
	r = ReturnError(err)
	a.Equal(r.message, "").
		Equal(r.typ, typeOK)
}
