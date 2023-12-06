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
	packages.NeedImports | packages.NeedModule | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo

const Cancelled = web.StringPhrase("cancelled")

type Packages struct {
	pkgsM sync.Mutex
	pkgs  []*packages.Package
	Fset  *token.FileSet
	l     *logger.Logger
}

func New(l *logger.Logger) *Packages {
	return &Packages{
		pkgs: make([]*packages.Package, 0, 10),
		Fset: token.NewFileSet(),
		l:    l,
	}
}

// ScanDir 添加 root 下的内容
//
// 仅在调用 [Parser.Parse] 之前添加有效果。
// root 添加的目录；
func (pkgs *Packages) ScanDir(ctx context.Context, root string, recursive bool) {
	root = filepath.Clean(root)

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

				ps, err := packages.Load(&packages.Config{
					Mode:    mode,
					Context: ctx,
					Dir:     dir,
					Fset:    pkgs.Fset,
				})
				if err != nil {
					pkgs.l.Error(err, "", 0)
					return
				}
				pkgs.append(ps...)
			}(dir)
		}
	}
	wg.Wait()
}

func (pkgs *Packages) append(ps ...*packages.Package) {
	pkgs.pkgsM.Lock()
	defer pkgs.pkgsM.Unlock()

	for _, p := range ps {
		if slices.IndexFunc(pkgs.pkgs, func(e *packages.Package) bool { return p.PkgPath == e.PkgPath }) < 0 {
			pkgs.pkgs = append(pkgs.pkgs, p)
		}
	}
}

func (pkgs *Packages) Package(path string) *packages.Package {
	pkg, _ := sliceutil.At(pkgs.pkgs, func(p *packages.Package, _ int) bool { return p.PkgPath == path })
	return pkg
}

func (pkgs *Packages) Position(p token.Pos) token.Position {
	return pkgs.Fset.Position(p)
}

// f 如果返回了 false，将退出循环
func (pkgs *Packages) Range(f func(*packages.Package) bool) {
	for _, p := range pkgs.pkgs {
		if !f(p) {
			return
		}
	}
}
