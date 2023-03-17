// SPDX-License-Identifier: MIT

package filter

import (
	"strings"
	"testing"

	"github.com/issue9/assert/v3"
)

func trimLeft(v *string) { *v = strings.TrimLeft(*v, " ") }

func trimRight(v *string) { *v = strings.TrimRight(*v, " ") }

func TestSanitizers(t *testing.T) {
	a := assert.New(t, false)

	s := Sanitizers(trimLeft, trimRight)
	str := "  str  "
	s(&str)
	a.Equal(str, "str")
}
