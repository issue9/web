// SPDX-License-Identifier: MIT

package pkg

import (
	"context"
	"fmt"
	"go/ast"
	"go/build"
	"go/types"
	"path"
	"slices"
	"strconv"
	"strings"
)

// 拆分 path 中的类型信息
func splitTypes(path string) (wrap func(types.Type) types.Type, p string) {
	funcs := []func(types.Type) types.Type{}

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
			t, err := pkgs.TypeOf(ctx, p)
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
		p := strings.Trim(i.Path.Value, `"`)
		var name string
		if i.Name == nil {
			name = path.Base(p)
		} else {
			name = i.Name.Name
		}

		if name == x.Name {
			return p + "." + expr.Sel.Name
		}
	}

	pos := pkgs.fset.Position(f.Pos())
	panic(fmt.Sprintf("语法错误：在 %s 中找不到 %s 对应的导入项", pos.Filename, x.Name))
}

// 如果是内置类型，那么返回参数中的 TypeSpec 为 nil；
// 返回的 bool 表示是否找到了 path 对应的类型；
func (pkgs *Packages) lookup(ctx context.Context, typePath string) (types.Object, *ast.TypeSpec, *ast.File, bool) {
	if typePath == "" {
		panic("参数 path 不能为空")
	}

	var pkgPath string
	typeName := typePath
	if index := strings.LastIndexByte(typePath, '.'); index >= 0 {
		pkgPath = typePath[:index]
		typeName = typePath[index+1:]
	}

	if o, ts, f, found := findInPkgs(pkgs.pkgs, pkgPath, typeName); found {
		return o, ts, f, true
	}

	// 出于性能考虑并未加载依赖项，但是可能会依赖部分标准库的类型，
	// 此处对标准库作了特殊处理：未找到标准库中的对象时会加载相应的包。
	if pkgPath != "" && strings.IndexByte(pkgPath, '.') < 0 {
		ps, err := pkgs.load(ctx, path.Join(build.Default.GOROOT, "src", pkgPath))
		if err != nil {
			pkgs.l.Error(err, "", 0)
			return nil, nil, nil, false
		}

		return findInPkgs(ps, pkgPath, typeName)
	}

	return nil, nil, nil, false
}
