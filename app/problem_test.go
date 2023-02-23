// SPDX-License-Identifier: MIT

package app

import (
	"testing"

	"github.com/issue9/assert/v3"
)

func TestProblem_sanitize(t *testing.T) {
	a := assert.New(t, false)

	var p *Problem
	ps, err := p.sanitize()
	a.NotError(err).
		Nil(ps)

	p = &Problem{
		Builder:  "rfc7807",
		IDPrefix: "abc#",
	}
	ps, err = p.sanitize()
	a.NotError(err).
		NotNil(ps).
		NotNil(ps.Builder).
		Equal(ps.IDPrefix, "abc#")

	p = &Problem{
		Builder:  "not-exists",
		IDPrefix: "abc#",
	}
	ps, err = p.sanitize()
	a.Nil(ps).Error(err).Equal(err.Field, "builder")
}