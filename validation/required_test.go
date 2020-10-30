// SPDX-License-Identifier: MIT

package validation

import (
	"testing"

	"github.com/issue9/assert"
)

func TestRequired(t *testing.T) {
	a := assert.New(t)
	val := 5

	r := Required("msg", false)
	a.Equal(r.Validate(0), "msg")
	a.Equal(r.Validate(nil), "msg")
	a.Equal(r.Validate(""), "msg")
	a.Equal(r.Validate([]string{}), "msg")
	a.Empty(r.Validate([]string{""}))
	a.Empty(r.Validate(&val))

	r = Required("msg", true)
	a.Equal(r.Validate(0), "msg")
	a.Empty(r.Validate(nil))
	a.Equal(r.Validate(""), "msg")
	a.Equal(r.Validate([]string{}), "msg")
	a.Empty(r.Validate([]string{""}))
	a.Empty(r.Validate(&val))
}
