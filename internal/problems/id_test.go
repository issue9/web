// SPDX-License-Identifier: MIT

package problems

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v3"
)

func TestStatus(t *testing.T) {
	a := assert.New(t, false)
	a.Equal(Status(ProblemBadGateway), http.StatusBadGateway)
	a.Zero(Status("not-exists"))
}
