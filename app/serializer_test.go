// SPDX-License-Identifier: MIT

package app

import (
	"testing"

	"github.com/issue9/assert/v3"
)

func TestRegisterMimetype(t *testing.T) {
	a := assert.New(t, false)

	a.NotNil(mimetypesFactory["json"].Marshal)
	RegisterMimetype(nil, nil, "json")
	a.Nil(mimetypesFactory["json"].Marshal)
}
