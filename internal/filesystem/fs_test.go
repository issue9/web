// SPDX-License-Identifier: MIT

package filesystem

import (
	"embed"
	"testing"

	"github.com/issue9/assert"
)

//go:embed filesystem.go
var f1 embed.FS

//go:embed filesystem_test.go
var f2 embed.FS

func TestMultipleFS(t *testing.T) {
	a := assert.New(t)

	m := MultipleFS(f1, f2)

	a.True(ExistsFS(m, "filesystem.go"))
	a.True(ExistsFS(m, "filesystem_test.go"))
	a.False(ExistsFS(m, "not-exists.go"))
	a.False(ExistsFS(f1, "filesystem_test.go"))
	a.False(ExistsFS(f2, "filesystem.go"))
}
