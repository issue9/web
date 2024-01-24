// SPDX-License-Identifier: MIT

package app

import (
	"testing"

	"github.com/issue9/assert/v3"
)

func TestRegisterMimetype(t *testing.T) {
	a := assert.New(t, false)

	v, f := mimetypesFactory.get("json")
	a.True(f).
		NotNil(v).
		NotNil(v.marshal)

	RegisterMimetype(nil, nil, "json")
	v, f = mimetypesFactory.get("json")
	a.True(f).
		NotNil(v).
		Nil(v.marshal)
}
