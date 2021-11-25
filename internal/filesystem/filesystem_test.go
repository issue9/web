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
	a.False(Exists("./filesystem.go.not.exists"))
}

func TestExistsFS(t *testing.T) {
	a := assert.New(t, false)
	fsys := os.DirFS("./")

	a.True(ExistsFS(fsys, "filesystem.go"))
	a.False(ExistsFS(fsys, "filesystem.go.not.exists"))
}
