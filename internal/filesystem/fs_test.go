// SPDX-License-Identifier: MIT

package filesystem

import (
	"embed"
	"io/fs"
	"testing"

	"github.com/issue9/assert"
)

var _ fs.FS = &MultipleFS{}

//go:embed filesystem.go
var f1 embed.FS

//go:embed filesystem_test.go
var f2 embed.FS

func TestMultipleFS(t *testing.T) {
	a := assert.New(t)

	m := NewMultipleFS(f1, f2)

	a.True(ExistsFS(m, "filesystem.go"))
	a.True(ExistsFS(m, "filesystem_test.go"))
	a.False(ExistsFS(m, "not-exists.go"))
	a.False(ExistsFS(f1, "filesystem_test.go"))
	a.False(ExistsFS(f2, "filesystem.go"))
}
