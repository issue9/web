// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package problems

import (
	"testing"

	"github.com/issue9/assert/v6"
)

func TestProblem(t *testing.T) {
	a := assert.New(t, false)
	ps := New[render]()

	p := ps.NewProblem()
	a.NotNil(p)

	t.Run("WithParam", func(t *testing.T) {
		a := a.New(t)

		a.Empty(p.Params)
		p.WithParam("p1", "v1")
		p.WithParam("p1", "v2")
		a.Length(p.Params, 1).
			Equal(p.Params[0].Name, "p1").
			Equal(p.Params[0].Reasons, []string{"v1", "v2"})
	})

	t.Run("WithExtensions", func(t *testing.T) {
		a := a.New(t)

		a.Nil(p.Extensions)
		p.WithExtensions("5")
		a.Equal(p.Extensions, "5")

		p.WithExtensions(5)
		a.Equal(p.Extensions, 5)
	})

	t.Run("WithInstance", func(t *testing.T) {
		a := a.New(t)

		a.Empty(p.Instance)
		p.WithInstance("5")
		a.Equal(p.Instance, "5")

		p.WithInstance("5xxxx")
		a.Equal(p.Instance, "5xxxx")
	})

	t.Run("free", func(t *testing.T) {
		a := a.New(t)

		a.False(p.isFree)

		ps.Free(p)
		a.True(p.isFree)

		p = ps.NewProblem()
		a.False(p.isFree)
	})
}
