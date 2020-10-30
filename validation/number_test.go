// SPDX-License-Identifier: MIT

package validation

import (
	"math"
	"testing"

	"github.com/issue9/assert"
)

func TestRange(t *testing.T) {
	a := assert.New(t)

	a.Panic(func() {
		Range("msg", 100, 5)
	})

	r := Range("msg", 5, math.MaxInt16)
	a.Empty(r.Validate(5))
	a.Empty(r.Validate(5.1))
	a.Empty(r.Validate(math.MaxInt8))
	a.Equal(r.Validate(math.MaxInt32), "msg")
	a.Equal(r.Validate(-1), "msg")
	a.Equal(r.Validate(-1.1), "msg")
	a.Equal(r.Validate("5"), "msg")

	r = Min("msg", 6)
	a.Empty(r.Validate(6))
	a.Empty(r.Validate(10))
	a.Equal(r.Validate(5), "msg")

	r = Max("msg", 6)
	a.Empty(r.Validate(6))
	a.Equal(r.Validate(10), "msg")
	a.Empty(r.Validate(5))
	a.Empty(r.Validate(uint(5)))
}
