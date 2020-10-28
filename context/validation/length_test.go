// SPDX-License-Identifier: MIT

package validation

import (
	"testing"

	"github.com/issue9/assert"
)

func TestLength(t *testing.T) {
	a := assert.New(t)

	l := Length("msg", "invalid-type", 5, 7)
	a.Equal(l.Validate("123"), "msg")
	a.Equal(l.Validate([]byte("123")), "msg")
	a.Empty(l.Validate([]rune("12345")))
	a.Equal(l.Validate(&struct{}{}), "invalid-type")
}
