// SPDX-License-Identifier: MIT

package filesystem

import (
	"testing"

	"github.com/issue9/assert"
)

func TestExists(t *testing.T) {
	a := assert.New(t)

	a.True(Exists("./"))
	a.True(Exists("../filesystem"))
	a.True(Exists("../filesystem/filesystem.go"))
	a.False(Exists("../filesystem/not-exists"))
	a.False(Exists("./not-exists"))
}
