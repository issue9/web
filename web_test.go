// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"runtime"
	"strings"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/version"
)

func TestVersion(t *testing.T) {
	assert.True(t, version.SemVerValid(Version), "无效的版本号")
}

func TestMiniVersion(t *testing.T) {
	a := assert.New(t)

	goversion := strings.TrimPrefix(runtime.Version(), "go")
	a.NotEmpty(goversion)

	// tip 版本，不作检测
	if strings.HasPrefix(goversion, "devel ") {
		return
	}

	v, err := version.SemVerCompare(goversion, MinimumGoVersion)
	a.NotError(err).False(v < 0, "版本号太低")
}
