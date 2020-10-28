// SPDX-License-Identifier: MIT

package validation

import (
	"math"
	"testing"

	"github.com/issue9/assert"
)

func TestLength(t *testing.T) {
	a := assert.New(t)

	a.Panic(func() {
		Length("msg", "invalid-type", -1, -1)
	})

	l := Length("msg", "invalid-type", 5, 7)
	a.Equal(l.Validate("123"), "msg")
	a.Equal(l.Validate([]byte("123")), "msg")
	a.Empty(l.Validate([]rune("12345")))
	a.Equal(l.Validate(&struct{}{}), "invalid-type")

	l = MinLength("msg", "invalid", 6)
	a.Empty(l.Validate("123456"))
	a.Empty(l.Validate("12345678910"))
	a.Equal(l.Validate("12345"), "msg")

	l = MaxLength("msg", "invalid", 6)
	a.Empty(l.Validate("123456"))
	a.Equal(l.Validate("12345678910"), "msg")
	a.Empty(l.Validate("12345"))
}

func TestRange(t *testing.T) {
	a := assert.New(t)

	r := Range("msg", "invalid-type", 5, math.MaxInt16)
	a.Empty(r.Validate(5))
	a.Empty(r.Validate(5.1))
	a.Empty(r.Validate(math.MaxInt8))
	a.Equal(r.Validate(math.MaxInt32), "msg")
	a.Equal(r.Validate(-1), "msg")
	a.Equal(r.Validate(-1.1), "msg")
	a.Equal(r.Validate("5"), "invalid-type")

	r = Min("msg", "invalid", 6)
	a.Empty(r.Validate(6))
	a.Empty(r.Validate(10))
	a.Equal(r.Validate(5), "msg")

	r = Max("msg", "invalid", 6)
	a.Empty(r.Validate(6))
	a.Equal(r.Validate(10), "msg")
	a.Empty(r.Validate(5))
}
