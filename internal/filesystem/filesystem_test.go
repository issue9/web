// SPDX-License-Identifier: MIT

package filesystem

import (
	"os"
	"testing"

	"github.com/issue9/assert/v2"
)

func TestExists(t *testing.T) {
	a := assert.New(t, false)
	a.True(Exists("./filesystem.go"))
	a.False(Exists("./not-exists.go"))
}

func TestExistsFS(t *testing.T) {
	a := assert.New(t, false)
	a.True(ExistsFS(os.DirFS("./"), "filesystem.go"))
	a.False(ExistsFS(os.DirFS("./"), "not-exists.go"))
}
