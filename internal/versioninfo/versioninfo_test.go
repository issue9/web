// SPDX-License-Identifier: MIT

package versioninfo

import (
	"path/filepath"
	"testing"

	"github.com/issue9/assert"
)

func TestRoot(t *testing.T) {
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

func TestDir_DumpFile(t *testing.T) {
	a := assert.New(t)

	v, err := Root("./testdata")
	a.NotError(err).NotNil(v)

	a.NotError(v.DumpFile())
	a.FileExists(filepath.Join(string(v), infoPath))
}

func TestParseDescribe(t *testing.T) {
	a := assert.New(t)

	tag, commits, hash := parseDescribe("v0.2.4-0-ge2f5e99a3306bba28e81f507bf66c905825184e5")
	a.Equal(tag, "0.2.4").
		Equal("0", commits).
		Equal(hash, "e2f5e99a3306bba28e81f507bf66c905825184e5")

	tag, commits, hash = parseDescribe("0.2.4-8-ge2f5e99a3306bba28e81")
	a.Equal(tag, "0.2.4").
		Equal("8", commits).
		Equal(hash, "e2f5e99a3306bba28e81")
}
