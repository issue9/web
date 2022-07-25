// SPDX-License-Identifier: MIT

package problem

import (
	"testing"

	"github.com/issue9/assert/v2"
)

var (
	_ Problem   = &rfc7807{}
	_ BuildFunc = RFC7807Builder
)

func TestProblem(t *testing.T) {
	a := assert.New(t, false)
	p := RFC7807Builder("id", "title", "detail", 400)
	a.NotNil(p)
	p.SetInstance("https://example.com/instance/1")

	pp, ok := p.(*rfc7807)
	a.True(ok).NotNil(pp)
	a.Equal(pp.Instance, "https://example.com/instance/1").
		Equal(pp.Status(), 400).
		Equal(pp.Type, "id")

	p.Destroy()
}
