// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package pkg

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io/fs"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/issue9/source"
	"golang.org/x/tools/go/packages"
)

// 拆分指定的字段类型
//
// type<field=type1,field2=type2<field1=type3>>
//
// 返回指向的类型以及需要被替换的字段列表。
func (pkgs *Packages) splitFieldTypes(ctx context.Context, path string) (p string, fieldTypes map[string]types.Type, err error) {
	index := strings.IndexByte(path, '<')
	if index < 0 {
		return path, nil, nil
	}

	if path[len(path)-1] != '>' {
		panic(fmt.Sprintf("%s 的最后一个字符必须得是 >", path))
	}

	p = path[:index]
	fieldTypes = make(map[string]types.Type, 3)

	appendTS := func(path string) error {
		ps := strings.SplitN(path, "=", 2)
		if len(ps) != 2 {
			panic(fmt.Sprintf("无效的语法 %s", path))
		}

		t, err := pkgs.typeOf(ctx, ps[1])
		if err != nil {
			return err
		}

		fieldTypes[ps[0]] = t
		return nil
	}

	var depth int
	start := index + 1
LOOP:
	for i := start; i < len(path); i++ {
		switch path[i] {
		case '<':
			depth++
		case '>':
			depth--
		case ',':
			if depth > 0 {
				continue LOOP
			}
			if err := appendTS(path[start:i]); err != nil {
				return "", nil, err
			}

			i++ // 忽略当前字符
			start = i
		}
	}

	if err := appendTS(path[start : len(path)-1]); err != nil {
		return "", nil, err
	}

	return p, fieldTypes, nil
}

// 拆分 path 中表示类似的前缀，比如 [] 表示数组
func splitTypes(path string) (wrap func(types.Type) types.Type, p string) {
	funcs := make([]func(types.Type) types.Type, 0, 5)

LOOP:
	for path != "" {
		switch path[0] {
		case '*':
			funcs = append(funcs, func(t types.Type) types.Type { return types.NewPointer(t) })
			path = path[1:]
			continue LOOP
		case '[':
			if len(path) > 1 && path[1] == ']' {
				funcs = append(funcs, func(t types.Type) types.Type { return types.NewSlice(t) })
				path = path[2:]
				continue LOOP
			}

			if i := strings.IndexByte(path, ']'); i > 0 {
				if num := strings.TrimSpace(path[1:i]); num == "" {
					funcs = append(funcs, func(t types.Type) types.Type { return types.NewSlice(t) })
				} else {
					size, err := strconv.ParseInt(num, 10, 64)
					if err != nil || size < 0 { // [xy] 如果 xy 不是数值，表示这不是数组，直接忽略
						break LOOP
					}
					funcs = append(funcs, func(t types.Type) types.Type { return types.NewArray(t, size) })
				}

				path = path[i+1:]
				continue LOOP
			}
		default:
			break LOOP
		}
	}
	slices.Reverse(funcs)

	return func(t types.Type) types.Type {
		for _, ff := range funcs {
			t = ff(t)
		}
		return t
	}, path
}

// 拆分 path 中的范型参数
func (pkgs *Packages) splitTypeParams(ctx context.Context, path string) (p string, tl typeList, err error) {
	if path = strings.TrimSpace(path); path == "" {
		return
	}

	var tps []string
	if index := strings.LastIndexByte(path, '['); index > 0 {
		tps = strings.Split(path[index+1:len(path)-1], ",")
		for k, v := range tps {
			tps[k] = strings.TrimSpace(v)
		}

		path = path[:index]
	}

	if len(tps) > 0 {
		ts := make([]types.Type, 0, len(tps))
		for _, p := range tps {
			t, err := pkgs.typeOf(ctx, p)
			if err != nil {
				return "", nil, err
			}
			ts = append(ts, t)
		}

		tl = newTypeList(ts...)
	}

	return path, tl, nil
}

func (pkgs *Packages) getPathFromSelectorExpr(expr *ast.SelectorExpr, f *ast.File) string {
	x, ok := expr.X.(*ast.Ident)
	if !ok {
		panic(fmt.Sprintf("expr.X 不是 ast.Ident 类型，而是 %T", expr.X))
	}

	for _, i := range f.Imports {
		raw := strings.Trim(i.Path.Value, `"`)
		p, ok := filterVersionSuffix(raw, '/')
		if !ok {
			p, _ = filterVersionSuffix(p, '.')
		}

		var name string
		if i.Name == nil {
			name = path.Base(p)
		} else {
			name = i.Name.Name
		}

		if name == x.Name {
			return raw + "." + expr.Sel.Name
		}
	}

	pos := pkgs.fset.Position(f.Pos())
	panic(fmt.Sprintf("语法错误：在 %s 中找不到 %s 对应的导入项", pos.Filename, x.Name))
}

// github.com/issue9/logs/v7 过滤掉 /v7
func filterVersionSuffix(p string, separator byte) (string, bool) {
	if index := strings.LastIndexByte(p, separator); index > 0 {
		if v := p[index+1:]; len(v) > 0 && v[0] == 'v' {
			isNumber := true
			for _, c := range v[1:] {
				if isNumber = c > '0' && c < '9'; !isNumber {
					break
				}
			}

			if isNumber {
				return p[:index], true
			}
		}
	}

	return p, false
}

// 如果是内置类型，那么返回参数中的 TypeSpec 为 nil；
// 返回的 bool 表示是否找到了 path 对应的类型；
// typePath 的路径部分应该是真实的路径而不是 Import 中的别名；
func (pkgs *Packages) lookup(ctx context.Context, typePath string) (types.Object, *ast.TypeSpec, *ast.File, bool) {
	if typePath == "" {
		panic("参数 path 不能为空")
	}

	var pkgPath string
	typeName := typePath
	// 防止出现 github.com/pkg/pkg.type/name 等不规则内容，type/name 并不是一个合法的类型名
	if index := strings.LastIndexByte(typePath, '.'); index > 0 && !strings.ContainsRune(typePath[index:], '/') {
		pkgPath = typePath[:index]
		typeName = typePath[index+1:]
	}

	if o, ts, f, found := findInPkgs(pkgs.pkgs, pkgPath, typeName); found {
		return o, ts, f, true
	}

	// 碰到未找到的类型，尝试加载依赖包。

	for pkgDir := range pkgs.pkgs {
		dir, err := source.PkgSourceDir(pkgPath, pkgDir, true)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, nil, nil, false
		}

		p, err := pkgs.load(ctx, dir)
		if err != nil {
			pkgs.l.Error(err, "", 0)
			return nil, nil, nil, false
		}

		if p != nil {
			return findInPkgs(map[string]*packages.Package{dir: p}, pkgPath, typeName)
		}
	}

	return nil, nil, nil, false
}

func findInPkgs(ps map[string]*packages.Package, pkgPath, typeName string) (types.Object, *ast.TypeSpec, *ast.File, bool) {
	for _, p := range ps {
		if p.PkgPath != pkgPath {
			continue
		}

		obj := p.Types.Scope().Lookup(typeName)
		if obj == nil {
			break // 不可能存在多个 path 相同但内容不同的 Package 对象
		}

		for _, f := range p.Syntax {
			for _, decl := range f.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok || gen.Tok != token.TYPE {
					continue
				}

				for _, spec := range gen.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok || ts.Name.Name != typeName {
						continue
					}

					// 整个 type() 范围只有一个类型，直接采用 type 的注释
					if ts.Doc == nil && len(gen.Specs) == 1 {
						ts.Doc = gen.Doc
					}
					return obj, ts, f, true
				}
			}
		}
	}

	return nil, nil, nil, false
}
