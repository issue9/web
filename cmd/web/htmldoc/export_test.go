// SPDX-License-Identifier: MIT

package htmldoc

import (
	"testing"

	"github.com/issue9/assert/v3"
)

func TestGetName(t *testing.T) {
	a := assert.New(t, false)

	a.Equal(getName("name", "-"), "-")
	a.Equal(getName("name", "tag"), "tag")
	a.Equal(getName("name", ""), "name")
}
