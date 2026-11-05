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
