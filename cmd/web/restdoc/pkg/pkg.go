// SPDX-License-Identifier: MIT

// Package pkg 用于对包的解析管理
package pkg

import (
	"context"
	"go/token"
	"path/filepath"
	"sync"

	"github.com/issue9/web"
	"golang.org/x/tools/go/packages"

	"github.com/issue9/web/cmd/web/restdoc/logger"
)

const mode = packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
	packages.NeedImports | packages.NeedModule | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo

const Cancelled = web.StringPhrase("cancelled")

type AppendFunc = func(...*packages.Package)

// ScanDir 添加 root 下的内容
//
// 仅在调用 [Parser.Parse] 之前添加有效果。
// root 添加的目录；
func ScanDir(ctx context.Context, fset *token.FileSet, root string, recursive bool, af AppendFunc, l *logger.Logger) {
	root = filepath.Clean(root)

	dirs, err := getDirs(root, recursive)
	if err != nil {
		l.Error(err, "", 0)
		return
	}

	wg := &sync.WaitGroup{}
	for _, dir := range dirs {
		select {
		case <-ctx.Done():
			l.Warning(Cancelled)
			return
		default:
			wg.Add(1)
			go func(dir string) {
				defer wg.Done()
				if p := scan(ctx, fset, l, dir); p != nil {
					af(p...)
				}
			}(dir)
		}
	}
	wg.Wait()
}

// 扫描 dir 下的 go 文件
//
// 不包含子目录和测试文件；
// modPath 为 dir 下 go 文件的导出路径；
func scan(ctx context.Context, fset *token.FileSet, l *logger.Logger, dir string) []*packages.Package {
	cfg := &packages.Config{
		Mode:    mode,
		Context: ctx,
		Dir:     dir,
		Fset:    fset,
	}
	pkgs, err := packages.Load(cfg)
	if err != nil {
		l.Error(err, "", 0)
		return nil
	}

	return pkgs
}
