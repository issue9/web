// SPDX-License-Identifier: MIT

package pkg

import (
	"context"
	"go/token"
	"sync"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web/logs"

	"github.com/issue9/web/cmd/web/internal/restdoc/logger/loggertest"
)

type appender struct {
	pkgs []*Package
	mux  sync.Mutex
}

func (a *appender) append(p *Package) {
	a.mux.Lock()
	defer a.mux.Unlock()
	a.pkgs = append(a.pkgs, p)
}

func newAppender() *appender {
	return &appender{pkgs: make([]*Package, 0, 10)}
}

func TestScanDir(t *testing.T) {
	a := assert.New(t, false)

	ap := newAppender()
	fset := token.NewFileSet()
	l := loggertest.New(a)
	ScanDir(context.Background(), fset, "./testdir", true, ap.append, l.Logger)
	a.Length(ap.pkgs, 2).
		Length(l.Records[logs.Info], 2).
		NotNil(ap.pkgs[0].Path, "github.com/issue9/web/cmd/web/internal/restdoc/pkg/testdir").
		NotNil(ap.pkgs[1].Path, "github.com/issue9/web/cmd/web/internal/restdoc/pkg/testdir/testdir2")

	ap = newAppender()
	fset = token.NewFileSet()
	l = loggertest.New(a)
	ScanDir(context.Background(), fset, "./testdir", false, ap.append, l.Logger)
	a.Length(ap.pkgs, 1).
		Length(l.Records[logs.Info], 2).
		NotNil(ap.pkgs[0].Path, "github.com/issue9/web/cmd/web/internal/restdoc/pkg/testdir/testdir2")
}

func TestScan(t *testing.T) {
	a := assert.New(t, false)

	l := loggertest.New(a)
	ctx := context.Background()
	fset := token.NewFileSet()

	pkg := scan(ctx, fset, l.Logger, "./testdir", "github.com/test/testdata")
	a.NotNil(pkg).
		Length(pkg.Files, 1).
		Equal(pkg.Path, "github.com/test/testdata").
		Zero(l.Count())
}
