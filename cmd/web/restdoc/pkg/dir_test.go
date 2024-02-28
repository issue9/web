// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package pkg

import (
	"testing"

	"github.com/issue9/assert/v4"
)

func TestGetDirs(t *testing.T) {
	a := assert.New(t, false)

	dirs, err := getDirs("./testdir", false)
	a.NotError(err).Length(dirs, 1)

	dirs, err = getDirs("./testdir", true)
	a.NotError(err).Length(dirs, 2)
}
