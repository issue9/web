// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

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
	path, err := FindRoot("./")
	a.NotError(err).Equal(path, abs)

	abs, err = filepath.Abs("./testdata")
	a.NotError(err)
	path, err = FindRoot("./testdata")
	a.NotError(err).Equal(path, abs)

	// 该目录不存在 go.mod
	path, err = FindRoot("./../../../../")
	a.Error(err).Empty(path)
}

func TestVarPath(t *testing.T) {
	a := assert.New(t)

	p, err := VarPath("./testdata")
	a.NotError(err)
	a.Equal(p, "testdata/v2/internal/version")

	p, err = VarPath("./")
	a.NotError(err)
	a.Equal(p, "github.com/issue9/web/internal/version")

	// 不存在 go.mod
	p, err = VarPath("../../../")
	a.Error(err).Empty(p)
}
