// SPDX-License-Identifier: MIT

package filesystem

import (
	"os"
	"testing"

	"github.com/issue9/assert"
)

func TestExists(t *testing.T) {
	a := assert.New(t)

	a.True(Exists("./filesystem.go"))
	a.False(Exists("./filesystem.go.not.exists"))
}

func TestExistsFS(t *testing.T) {
	a := assert.New(t)
	fsys := os.DirFS("./")

	a.True(ExistsFS(fsys, "filesystem.go"))
	a.False(ExistsFS(fsys, "filesystem.go.not.exists"))
}
