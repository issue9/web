// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package htmldoc

import (
	"testing"

	"github.com/issue9/assert/v4"
)

func TestGetName(t *testing.T) {
	a := assert.New(t, false)

	a.Equal(getName("name", "-"), "-").
		Equal(getName("name", "tag"), "tag").
		Equal(getName("name", ""), "name")
}
