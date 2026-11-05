// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package sanitizers

import (
	"math"
	"testing"

	"github.com/issue9/assert/v6"
)

func TestAbs(t *testing.T) {
	a := assert.New(t, false)

	f := 5.1
	Abs(&f)
	a.Equal(f, 5.1)

	f = -5.1
	Abs(&f)
	a.Equal(f, 5.1)

	num := 5
	Abs(&num)
	a.Equal(num, 5)

	num = -5
	Abs(&num)
	a.Equal(num, 5)

	inf := math.Inf(1)
	Abs(&inf)
	a.Equal(inf, math.Inf(1))

	inf = math.Inf(-1)
	Abs(&inf)
	a.Equal(inf, math.Inf(1))

}
