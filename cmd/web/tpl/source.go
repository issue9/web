// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

package tpl

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/issue9/errwrap"
	"golang.org/x/mod/modfile"
)

type replacer struct {
	start, end int    // 需要替换的起止位置
	bytes      []byte // 替换的字符串
}

func replace(data []byte, replaces []replacer) []byte {
	buf := &errwrap.Buffer{}

	start := 0
	for _, item := range replaces {
		buf.WBytes(data[start:item.start]).WBytes(item.bytes)
		start = item.end + 1
	}
	if start <= len(data) {
		buf.WBytes(data[start:])
	}

	return buf.Bytes()
}

func replaceGoSourcePackageName(fset *token.FileSet, filename string, data []byte, newName string) ([]byte, error) {
	f, err := parser.ParseFile(fset, filename, data, parser.DeclarationErrors)
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(filename, "_test.go") && strings.HasSuffix(f.Name.Name, "_test") {
		r := replacer{
			start: fset.Position(f.Name.Pos()).Offset,
			end:   fset.Position(f.Name.End()).Offset - 1,
			bytes: []byte(newName + "_test"),
		}

		return replace(data, []replacer{r}), nil
	}

	if f.Name.Name != newName {
		r := replacer{
			start: fset.Position(f.Name.Pos()).Offset,
			end:   fset.Position(f.Name.End()).Offset - 1,
			bytes: []byte(newName),
		}
		return replace(data, []replacer{r}), nil
	}

	return data, nil
}

// 替换 Go 源码中的 import 语句
func replaceGoSourceImport(fset *token.FileSet, filename string, data []byte, oldPath, newPath string) ([]byte, error) {
	f, err := parser.ParseFile(fset, filename, data, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	items := make([]replacer, 0, 10)
	for _, spec := range f.Imports {
		old, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, err
		}

		if old == oldPath {
			if spec.Name == nil { // 可能还需要改包名
				items = append(items, replacer{
					start: fset.Position(spec.Path.Pos()).Offset,
					end:   fset.Position(spec.Path.End()).Offset - 1,
					bytes: []byte(path.Base(old) + " " + strconv.Quote(newPath)),
				})
			}
		} else if strings.HasPrefix(old, oldPath) {
			items = append(items, replacer{
				start: fset.Position(spec.Path.Pos()).Offset,
				end:   fset.Position(spec.Path.End()).Offset - 1,
				bytes: []byte(strconv.Quote(newPath + old[len(oldPath):])),
			})
		}
	}

	return replace(data, items), nil
}

func replaceGo(dir, oldPath, newPath string) error {
	if err := replaceGoPackageName(dir, filepath.Base(newPath)); err != nil {
		return err
	}

	if err := replaceGoImport(dir, oldPath, newPath); err != nil {
		return err
	}

	return replaceGoMod(filepath.Join(dir, "go.mod"), newPath)
}

// 替换 go.mod 中的 module 语句
func replaceGoMod(file string, newPath string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	f, err := modfile.ParseLax(file, data, nil)
	if err != nil {
		return err
	}
	if err := f.AddModuleStmt(newPath); err != nil {
		return err
	}

	data, err = f.Format()
	if err != nil {
		return err
	}
	return os.WriteFile(file, data, os.ModePerm)
}

// 将 dir 目录下的所有 Go 源码文件中的包名从 oldName 替换为 newName
func replaceGoPackageName(dir, newName string) error {
	items, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	fset := token.NewFileSet()
	for _, item := range items {
		if item.IsDir() {
			continue
		}

		if !strings.HasSuffix(item.Name(), ".go") {
			continue
		}

		p := filepath.Join(dir, item.Name())
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}

		data, err = replaceGoSourcePackageName(fset, p, data, newName)
		if err != nil {
			return err
		}

		if err := os.WriteFile(p, data, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

// 替换 import 中对项目自身的引用
func replaceGoImport(dir, oldPath, newPath string) error {
	fset := token.NewFileSet()

	return filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || strings.ToLower(filepath.Ext(d.Name())) != ".go" {
			return nil
		}

		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}

		data, err = replaceGoSourceImport(fset, p, data, oldPath, newPath)
		if err != nil {
			return err
		}

		return os.WriteFile(p, data, os.ModePerm)
	})
}
