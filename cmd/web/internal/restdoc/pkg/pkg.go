// SPDX-License-Identifier: MIT

// Package pkg 用于对包的解析管理
package pkg

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/issue9/localeutil"
	"github.com/issue9/source"
	"github.com/issue9/web"

	"github.com/issue9/web/cmd/web/internal/restdoc/logger"
)

type Package struct {
	Path  string // 当前包的 path
	Files []*ast.File
}

type AppendFunc = func(*Package)

// AddDir 添加 root 下的内容
//
// 仅在调用 [RESTDoc.Openapi3] 之前添加有效果。
// root 添加的目录；
func ScanDir(ctx context.Context, fset *token.FileSet, root string, recursive bool, af AppendFunc, l *logger.Logger) {
	root = filepath.Clean(root)

	l.Info(localeutil.Phrase("start parse %s ...\n", root))

	dirs, err := getDirs(root, recursive)
	if err != nil {
		l.Error(err, "", 0)
		return
	}

	modPath, err := source.ModPath(root)
	if err != nil {
		l.Error(err, "", 0)
		return
	}

	wg := &sync.WaitGroup{}
	for _, dir := range dirs {
		select {
		case <-ctx.Done():
			l.Warning(web.Phrase("cancelled"))
			return
		default:
			wg.Add(1)
			go func(dir string) {
				defer wg.Done()

				suffix := strings.TrimPrefix(filepath.Clean(dir), root)
				suffix = strings.TrimFunc(suffix, func(r rune) bool { return r == filepath.Separator })
				p := scan(ctx, fset, l, dir, path.Join(modPath, suffix))
				if p != nil {
					af(p)
				}
			}(dir)
		}
	}
	wg.Wait()
	l.Info(localeutil.Phrase("parse %s complete\n", root))
}

// 扫描 dir 下的 go 文件
//
// 不包含子目录和测试文件；
// modPath 为 dir 下 go 文件的导出路径；
func scan(ctx context.Context, fset *token.FileSet, l *logger.Logger, dir, modPath string) *Package {
	entry, err := os.ReadDir(dir)
	if err != nil {
		l.Error(err, "", 0)
		return nil
	}

	astFiles := make([]*ast.File, 0, len(entry))
	astFilesM := &sync.Mutex{}
	appendFiles := func(f *ast.File) {
		astFilesM.Lock()
		defer astFilesM.Unlock()
		astFiles = append(astFiles, f)
	}

	wg := &sync.WaitGroup{}
	for _, e := range entry {
		select {
		case <-ctx.Done():
			l.Warning(web.Phrase("cancelled"))
			return nil
		default:
			// 路径、非 .go 扩展名 或是 _test.go 结尾的文件都忽略
			name := strings.ToLower(e.Name())
			if e.IsDir() || filepath.Ext(name) != ".go" || strings.HasSuffix(name, "_test.go") {
				continue
			}

			wg.Add(1)
			go func(path string) {
				defer wg.Done()

				f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
				if err == nil {
					appendFiles(f)
					return
				}
				l.Error(err, "", 0)
			}(filepath.Join(dir, e.Name()))
		}
	}
	wg.Wait()

	return &Package{
		Path:  modPath,
		Files: astFiles,
	}
}
