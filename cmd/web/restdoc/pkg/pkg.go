// SPDX-License-Identifier: MIT

// Package pkg 用于对包的解析管理
package pkg

import (
	"context"
	"go/token"
	"path/filepath"
	"slices"
	"sync"

	"github.com/issue9/sliceutil"
	"github.com/issue9/web"
	"golang.org/x/tools/go/packages"

	"github.com/issue9/web/cmd/web/restdoc/logger"
)

const mode = packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
	packages.NeedModule | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo

const Cancelled = web.StringPhrase("cancelled")

// Packages 管理加载的包
type Packages struct {
	pkgsM sync.Mutex
	pkgs  []*packages.Package
	fset  *token.FileSet
	l     *logger.Logger
}

func New(l *logger.Logger) *Packages {
	return &Packages{
		pkgs: make([]*packages.Package, 0, 10),
		fset: token.NewFileSet(),
		l:    l,
	}
}

// ScanDir 添加 root 下的内容
//
// root 添加的目录；
func (pkgs *Packages) ScanDir(ctx context.Context, root string, recursive bool) {
	root = filepath.Clean(root)

	pkgs.l.Info(web.Phrase("scan source dir %s", root))

	dirs, err := getDirs(root, recursive)
	if err != nil {
		pkgs.l.Error(err, "", 0)
		return
	}

	wg := &sync.WaitGroup{}
	for _, dir := range dirs {
		select {
		case <-ctx.Done():
			pkgs.l.Warning(Cancelled)
			return
		default:
			wg.Add(1)
			go func(dir string) {
				defer wg.Done()

				if _, err := pkgs.load(ctx, dir); err != nil {
					pkgs.l.Error(err, "", 0)
					return
				}
			}(dir)
		}
	}
	wg.Wait()
}

func (pkgs *Packages) load(ctx context.Context, dir string) ([]*packages.Package, error) {
	ps, err := packages.Load(&packages.Config{
		Mode:    mode,
		Context: ctx,
		Dir:     dir,
		Fset:    pkgs.fset,
	})
	if err != nil {
		return nil, err
	}

	pkgs.pkgsM.Lock()
	defer pkgs.pkgsM.Unlock()
	for _, p := range ps {
		if slices.IndexFunc(pkgs.pkgs, func(e *packages.Package) bool { return p.PkgPath == e.PkgPath }) < 0 {
			pkgs.pkgs = append(pkgs.pkgs, p)
		}
	}

	return ps, nil
}

func (pkgs *Packages) FileSet() *token.FileSet { return pkgs.fset }

// Package 返回指定路径的包对象
func (pkgs *Packages) Package(path string) *packages.Package {
	pkg, _ := sliceutil.At(pkgs.pkgs, func(p *packages.Package, _ int) bool { return p.PkgPath == path })
	return pkg
}

// Range 依次访问已经加载的包
//
// f 如果返回了 false，将退出循环
func (pkgs *Packages) Range(f func(*packages.Package) bool) {
	for _, p := range pkgs.pkgs {
		if !f(p) {
			return
		}
	}
}
