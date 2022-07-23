// SPDX-License-Identifier: MIT

package serializer

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v2"
)

var (
	_ Problem = &StandardsProblem{}
	_ Problem = &InvalidParamsProblem{}
)

func TestProblem(t *testing.T) {
	a := assert.New(t, false)
	p := NewStandardsProblem()
	a.NotNil(p)

	p.SetType("https://example.com/problem/400")
	a.Equal(p.GetType(), "https://example.com/problem/400")

	p.SetTitle("title")
	a.Equal(p.GetTitle(), "title")

	p.SetDetail("detail")
	a.Equal(p.GetDetail(), "detail")

	p.SetStatus(http.StatusOK)
	a.Equal(p.GetStatus(), http.StatusOK)

	p.SetInstance("https://example.com/instance/1")
	a.Equal(p.GetInstance(), "https://example.com/instance/1")

	p.Destroy()

	p = NewStandardsProblem()
	a.Equal(p.GetType(), "")
}
