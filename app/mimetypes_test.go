// SPDX-License-Identifier: MIT

package app

import (
	"testing"

	"github.com/issue9/assert/v2"
)

func TestRegisterMimetype(t *testing.T) {
	a := assert.New(t, false)
	a.Panic(func() {
		RegisterMimetype(nil, nil)
	})

	a.NotNil(mimetypesFactory["application/json"].m)
	RegisterMimetype(nil, nil, "application/json")
	a.Nil(mimetypesFactory["application/json"].m)
}
