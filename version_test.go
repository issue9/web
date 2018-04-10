// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/version"
)

func TestVersion(t *testing.T) {
	assert.True(t, version.SemVerValid(Version), "无效的版本号")
}
