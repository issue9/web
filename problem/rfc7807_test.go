// SPDX-License-Identifier: MIT

package problem

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v2"
)

var _ Problem = &RFC7807{}

func TestProblem(t *testing.T) {
	a := assert.New(t, false)
	p := NewRFC7807(nil)
	a.NotNil(p)

	p.SetType("https://example.com/problem/400")
	a.Equal(p.Type, "https://example.com/problem/400")

	p.SetTitle("title")
	a.Equal(p.Title, "title")

	p.SetDetail("detail")
	a.Equal(p.Detail, "detail")

	p.SetStatus(http.StatusOK)
	a.Equal(p.Status, http.StatusOK)

	p.SetInstance("https://example.com/instance/1")
	a.Equal(p.Instance, "https://example.com/instance/1")

	p.Destroy()

	p = NewRFC7807(nil)
	a.Equal(p.Type, "")
}
