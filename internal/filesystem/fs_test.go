// SPDX-License-Identifier: MIT

package filesystem

import (
	"embed"
	"io/fs"
	"os"
	"testing"

	"github.com/issue9/assert/v2"
)

var _ fs.FS = &MultipleFS{}

//go:embed filesystem.go
var f1 embed.FS

//go:embed filesystem_test.go
var f2 embed.FS

func TestMultipleFS(t *testing.T) {
	a := assert.New(t, false)

	m := NewMultipleFS(f1, f2)
	a.NotNil(m)

	a.True(ExistsFS(m, "filesystem.go"))
	a.True(ExistsFS(m, "filesystem_test.go"))
	a.False(ExistsFS(m, "not-exists.go"))
	a.False(ExistsFS(f1, "filesystem_test.go"))
	a.False(ExistsFS(f2, "filesystem.go"))
}

func TestMultipleFS_Glob(t *testing.T) {
	a := assert.New(t, false)

	f1 := os.DirFS("./")
	f2 := os.DirFS("./testdata")

	m := NewMultipleFS(f1)
	a.NotNil(m)
	matches, err := fs.Glob(m, "filesystem.go")
	a.NotError(err).Equal(matches, []string{"filesystem.go"})

	// f1 存在于 "./testdata"
	matches, err = fs.Glob(m, "f1.txt")
	a.NotError(err).Empty(matches)
	matches, err = fs.Glob(m, "testdata/f1.txt")
	a.NotError(err).Equal(matches, []string{"testdata/f1.txt"})
	m.Add(f2)
	matches, err = fs.Glob(m, "f1.txt")
	a.NotError(err).Equal(matches, []string{"f1.txt"})

	// fs_* 同时匹配多个 fs.FS
	matches, err = fs.Glob(m, "fs_*")
	a.NotError(err).Equal(matches, []string{"fs_test.go"})

	m = NewMultipleFS(f2, f1) // 调换顺序
	matches, err = fs.Glob(m, "fs_*")
	a.NotError(err).Equal(matches, []string{"fs_test.txt"})
}
