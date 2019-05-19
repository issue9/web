// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package release

import (
	"path/filepath"
	"testing"

	"github.com/issue9/assert"
)

func TestFindRoot(t *testing.T) {
	a := assert.New(t)

	abs, err := filepath.Abs("./../../..")
	a.NotError(err)
	path, err := FindRoot("./")
	a.NotError(err).Equal(path, abs)

	// 该目录不存在 go.mod
	path, err = FindRoot("./../../../../")
	a.Error(err).Empty(path)
}
