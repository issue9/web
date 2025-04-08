// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

package tpl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/issue9/assert/v4"
)

func TestDownload(t *testing.T) {
	a := assert.New(t, false)
	dir := "./testdata/out/download"

	a.NotError(download("github.com/issue9/web", dir)).
		FileExists(filepath.Join(dir, "LICENSE")).
		FileExists(filepath.Join(dir, "go.mod")).
		NotError(os.RemoveAll(dir))
}
