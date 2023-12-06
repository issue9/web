// SPDX-License-Identifier: MIT

package pkg

import (
	"context"
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/cmd/web/restdoc/logger/loggertest"
)

func TestScanDir(t *testing.T) {
	a := assert.New(t, false)
	l := loggertest.New(a)
	p := New(l.Logger)

	p.ScanDir(context.Background(), "./testdir", true)
	a.Length(p.pkgs, 2)

	l = loggertest.New(a)
	p = New(l.Logger)
	p.ScanDir(context.Background(), "./testdir", false)
	a.Length(p.pkgs, 1).
		Equal(p.pkgs[0].PkgPath, "github.com/issue9/web/restdoc/pkg")
}
