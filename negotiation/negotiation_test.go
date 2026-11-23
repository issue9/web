// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package negotiation

import (
	"testing"

	"github.com/issue9/assert/v6"
)

func TestBuildCompression(t *testing.T) {
	a := assert.New(t, false)

	c := buildCompression(nil, nil)
	a.True(c.wildcard).
		Length(c.types, 0).
		Length(c.wildcardSuffix, 0)

	c = buildCompression(nil, []string{"text"})
	a.Equal(c.types, []string{"text"})

	c = buildCompression(nil, []string{"text", "*"})
	a.Nil(c.types).
		True(c.wildcard).
		Nil(c.wildcardSuffix)
}
