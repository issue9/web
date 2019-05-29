// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package versioninfo

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/issue9/assert"
)

func TestFindRoot(t *testing.T) {
	a := assert.New(t)

	abs, err := filepath.Abs("../..")
	a.NotError(err)
	path, err := findRoot("./")
	a.NotError(err).Equal(path, abs)

	abs, err = filepath.Abs("./testdata")
	a.NotError(err)
	path, err = findRoot("./testdata")
	a.NotError(err).Equal(path, abs)

	// 该目录不存在 go.mod
	path, err = findRoot("./../../../../")
	a.Error(err).Empty(path)
}

func TestVersionInfo_DumpFile(t *testing.T) {
	a := assert.New(t)

	v, err := New("./testdata")
	a.NotError(err).NotNil(v)

	a.NotError(v.DumpFile("1.1.1"))
	a.FileExists(filepath.Join("./testdata/", Path))
}

func TestVersionInfoLDFlags(t *testing.T) {
	a := assert.New(t)
	now := time.Now().Format(buildDateLayout)

	v, err := New("./testdata")
	a.NotError(err).NotNil(v)
	p, err := v.LDFlags()
	a.NotError(err)
	a.Equal(p, fmt.Sprintf("-X testdata/v2/internal/version.buildDate=%s", now))

	v, err = New("./")
	a.NotError(err).NotNil(v)
	p, err = v.LDFlags()
	a.NotError(err)
	a.Equal(p, fmt.Sprintf("-X github.com/issue9/web/internal/version.buildDate=%s", now))
}
