// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package git

import (
	"strings"
	"testing"

	"github.com/issue9/assert/v4"
)

func TestVersion(t *testing.T) {
	a := assert.New(t, false)

	v, err := Version()
	a.NotError(err).NotEmpty(v)

	f, err := Commit(true)
	a.NotError(err).NotEmpty(f)

	s, err := Commit(false)
	a.NotError(err).NotEmpty(s).True(strings.HasPrefix(f, s), "v1=%s,v2=%s", f, s)
}
