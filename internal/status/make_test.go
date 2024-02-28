// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package status

import (
	"testing"

	"github.com/issue9/assert/v4"
)

func TestGetStatuses(t *testing.T) {
	a := assert.New(t, false)

	status, err := Get()
	a.NotError(err).NotEmpty(status).
		True(status[0].Value > 399)

	a.Equal(status[0].ID(), "ProblemBadRequest")
}
