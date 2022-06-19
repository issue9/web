// SPDX-License-Identifier: MIT

package app

import (
	"testing"

	"github.com/issue9/assert/v2"
)

func TestRegisterMimetype(t *testing.T) {
	a := assert.New(t, false)

	a.NotNil(mimetypesFactory["json"].m)
	RegisterMimetype(nil, nil, "json")
	a.Nil(mimetypesFactory["json"].m)
}
