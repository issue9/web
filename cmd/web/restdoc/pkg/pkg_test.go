// SPDX-License-Identifier: MIT

package pkg

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/cmd/web/restdoc/logger/loggertest"
)

func TestPackages_ScanDir(t *testing.T) {
	a := assert.New(t, false)
	l := loggertest.New(a)
	p := New(l.Logger)

	p.ScanDir(context.Background(), "./testdir", true)
	a.Length(p.pkgs, 2)

	l = loggertest.New(a)
	p = New(l.Logger)
	p.ScanDir(context.Background(), "./testdir", false)
	a.Length(p.pkgs, 1).
		Equal(p.pkgs[filepath.Clean("./testdir")].PkgPath, "github.com/issue9/web/restdoc/pkg")
}
