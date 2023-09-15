// SPDX-License-Identifier: MIT

package make

import (
	"testing"

	"github.com/issue9/assert/v3"
)

func TestGetStatuses(t *testing.T) {
	a := assert.New(t, false)

	status, err := GetStatuses()
	a.NotError(err).NotEmpty(status).
		True(status[0].Value > 399)

	a.Equal(status[0].ID(), "ProblemBadRequest")
}
