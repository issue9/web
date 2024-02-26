// SPDX-License-Identifier: MIT

// Package pkg 用于对包的解析管理
package pkg

import (
	"context"
	"fmt"
	"go/token"
	"path/filepath"
	"sync"

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
	pkgs  map[string]*packages.Package // 键名为对应的目录名
	fset  *token.FileSet
	l     *logger.Logger
}

func New(l *logger.Logger) *Packages {
	return &Packages{
		pkgs: make(map[string]*packages.Package, 30),
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

func (pkgs *Packages) load(ctx context.Context, dir string) (*packages.Package, error) {
	dir = filepath.Clean(dir)

	pkgs.pkgsM.Lock()
	defer pkgs.pkgsM.Unlock()
	if p, found := pkgs.pkgs[dir]; found {
		return p, nil
	}

	ps, err := packages.Load(&packages.Config{
		Mode:    mode,
		Context: ctx,
		Dir:     dir,
		Fset:    pkgs.fset,
	})
	if err != nil {
		return nil, err
	}

	switch len(ps) {
	case 0:
		return nil, nil
	case 1:
		pkgs.pkgs[dir] = ps[0]
		return ps[0], nil
	default:
		panic(fmt.Sprintf("目录 %s 中包的数量大于 1：%d", dir, len(ps)))
	}
}

func (pkgs *Packages) FileSet() *token.FileSet { return pkgs.fset }

// Package 返回指定路径的包对象
func (pkgs *Packages) Package(path string) *packages.Package {
	for _, p := range pkgs.pkgs {
		if p.PkgPath == path {
			return p
		}
	}
	return nil
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
