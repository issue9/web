// SPDX-License-Identifier: MIT

package versioninfo

import (
	"path/filepath"
	"testing"

	"github.com/issue9/assert"
)

func TestFindRoot(t *testing.T) {
	a := assert.New(t)

	abs, err := filepath.Abs("../..")
	a.NotError(err)
	path, err := Root("./")
	a.NotError(err).Equal(path, abs)

	abs, err = filepath.Abs("./testdata")
	a.NotError(err)
	path, err = Root("./testdata")
	a.NotError(err).Equal(path, abs)

	// 该目录不存在 go.mod
	path, err = Root("./../../../../")
	a.Error(err).Equal(path, Dir(""))
}

func TestDir_DumpVersionFile(t *testing.T) {
	a := assert.New(t)

	v, err := Root("./testdata")
	a.NotError(err).NotNil(v)

	a.NotError(v.DumpVersionFile("./"))
	a.FileExists(filepath.Join("./testdata/", versionPath))
}

func TestDir_DumpInfoFile(t *testing.T) {
	a := assert.New(t)

	v, err := Root("./testdata")
	a.NotError(err).NotNil(v)

	a.NotError(v.DumpInfoFile())
	a.FileExists(filepath.Join("./testdata/", infoPath))
}
