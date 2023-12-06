// SPDX-License-Identifier: MIT

package pkg

import (
	"context"
	"go/token"
	"sync"
	"testing"

	"github.com/issue9/assert/v3"
	"golang.org/x/tools/go/packages"

	"github.com/issue9/web/cmd/web/restdoc/logger/loggertest"
)

type appender struct {
	pkgs []*packages.Package
	mux  sync.Mutex
}

func (a *appender) append(p ...*packages.Package) {
	a.mux.Lock()
	defer a.mux.Unlock()
	a.pkgs = append(a.pkgs, p...)
}

func newAppender() *appender {
	return &appender{pkgs: make([]*packages.Package, 0, 10)}
}

func TestScanDir(t *testing.T) {
	a := assert.New(t, false)

	ap := newAppender()
	fset := token.NewFileSet()
	l := loggertest.New(a)
	ScanDir(context.Background(), fset, "./testdir", true, ap.append, l.Logger)
	a.Length(ap.pkgs, 2)

	ap = newAppender()
	fset = token.NewFileSet()
	l = loggertest.New(a)
	ScanDir(context.Background(), fset, "./testdir", false, ap.append, l.Logger)
	a.Length(ap.pkgs, 1).
		Equal(ap.pkgs[0].PkgPath, "github.com/issue9/web/restdoc/pkg")
}

func TestScan(t *testing.T) {
	a := assert.New(t, false)

	l := loggertest.New(a)
	ctx := context.Background()
	fset := token.NewFileSet()

	pkg := scan(ctx, fset, l.Logger, "./testdir")
	a.Length(pkg, 1).
		Equal(pkg[0].PkgPath, "github.com/issue9/web/restdoc/pkg").
		Zero(l.Count())

	pkg = scan(ctx, fset, l.Logger, "./testdir/testdir2")
	a.Length(pkg, 1).
		Equal(pkg[0].PkgPath, "github.com/issue9/web/restdoc/pkg/testdir2").
		Zero(l.Count())
}
