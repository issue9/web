// SPDX-FileCopyrightText: 2022-2026 caixw
//
// SPDX-License-Identifier: MIT

package sanitizers

import (
	"testing"

	"github.com/issue9/assert/v6"
)

func TestSanitizers(t *testing.T) {
	a := assert.New(t, false)

	s := Sanitizers(TrimLeft, TrimRight)
	str := "  str  "
	s(&str)
	a.Equal(str, "str")
}

func TestTrim(t *testing.T) {
	a := assert.New(t, false)

	s := " abc\t"
	Sanitizers(TrimLeft, TrimRight)(&s)
	a.Equal(s, "abc")

	s = " abc\t"
	Trim(&s)
	a.Equal(s, "abc")

	s = " ab\tc\t"
	Trim(&s)
	a.Equal(s, "ab\tc")

	s = " ab\tc\t"
	TrimLeft(&s)
	a.Equal(s, "ab\tc\t")

	s = " ab\tc\t"
	TrimRight(&s)
	a.Equal(s, " ab\tc")
}

func TestLower_Upper(t *testing.T) {
	a := assert.New(t, false)

	s := "Abc"
	Lower(&s)
	a.Equal(s, "abc")

	s = "Abc"
	Upper(&s)
	a.Equal(s, "ABC")
}
