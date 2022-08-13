// SPDX-License-Identifier: MIT

package server

import (
	"testing"
	"time"

	"github.com/issue9/assert/v3"
)

func TestSanitizeOptions(t *testing.T) {
	a := assert.New(t, false)

	o, err := sanitizeOptions(nil)
	a.NotError(err).NotNil(o)
	a.Equal(o.Location, time.Local).
		NotNil(o.Logs).
		NotNil(o.ProblemBuilder)
}
