// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package status

import (
	"testing"

	"github.com/issue9/assert/v6"
)

func TestGet(t *testing.T) {
	a := assert.New(t, false)

	status, err := Get()
	a.NotError(err).NotEmpty(status).
		Equal(status[0].Value, "400").
		Equal(status[0].ID(), "BadRequest")
}
