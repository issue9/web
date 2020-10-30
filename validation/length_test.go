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

	// 不限制长度
	l = Length("msg", "invalid-type", -1, -1)
	a.Empty(l.Validate("12345678910"))
	a.Empty(l.Validate([]rune("")))

	l = MinLength("msg", "invalid", 6)
	a.Empty(l.Validate("123456"))
	a.Empty(l.Validate("12345678910"))
	a.Equal(l.Validate("12345"), "msg")

	l = MaxLength("msg", "invalid", 6)
	a.Empty(l.Validate("123456"))
	a.Equal(l.Validate("12345678910"), "msg")
	a.Empty(l.Validate("12345"))
}
