// SPDX-License-Identifier: MIT

package schema

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/issue9/web/cmd/web/restdoc/openapi"
)

type typeParam struct {
	ref  *Ref
	name string // 类型参数的实参名称
}

func (s *Schema) fromIndexExprType(t *openapi.OpenAPI, file *ast.File, currPath, tag string, idx *ast.IndexExpr) (*Ref, error) {
	mod, idxName := getExprName(file, currPath, idx.Index)
	if mod != currPath {
		idxName = mod + "." + idxName
	}

	mod, name := getExprName(file, currPath, idx.X)
	return s.fromName(t, mod, name+"["+idxName+"]", tag, false)
}

func (s *Schema) fromIndexListExprType(t *openapi.OpenAPI, file *ast.File, currPath, tag string, idx *ast.IndexListExpr) (*Ref, error) {
	indexes := make([]string, 0, len(idx.Indices))
	for _, i := range idx.Indices {
		mod, idxName := getExprName(file, currPath, i)
		if mod != currPath {
			idxName = mod + "." + idxName
		}
		indexes = append(indexes, idxName)
	}

	mod, name := getExprName(file, currPath, idx.X)
	name += "[" + strings.Join(indexes, ",") + "]"
	return s.fromName(t, mod, name, tag, false)
}

func (s *Schema) fromIndexExpr(t *openapi.OpenAPI, file *ast.File, currPath, tag string, idx *ast.IndexExpr, refs map[string]*typeParam) (*Ref, error) {
	mod, idxName := getExprName(file, currPath, idx.Index)
	if mod != currPath {
		idxName = mod + "." + idxName
	} else {
		if elem, found := refs[idxName]; found {
			idxName = elem.name
		}
	}

	mod, name := getExprName(file, currPath, idx.X)
	return s.fromName(t, mod, name+"["+idxName+"]", tag, false)
}

func (s *Schema) fromIndexListExpr(t *openapi.OpenAPI, file *ast.File, currPath, tag string, idx *ast.IndexListExpr, refs map[string]*typeParam) (*Ref, error) {
	indexes := make([]string, 0, len(idx.Indices))
	for _, i := range idx.Indices {
		mod, idxName := getExprName(file, currPath, i)
		if mod != currPath {
			idxName = mod + "." + idxName
		} else {
			if elem, found := refs[idxName]; found {
				idxName = elem.name
			}
		}
		indexes = append(indexes, idxName)
	}

	mod, name := getExprName(file, currPath, idx.X)
	name += "[" + strings.Join(indexes, ",") + "]"
	return s.fromName(t, mod, name, tag, false)
}

func getExprName(file *ast.File, currPath string, expr ast.Expr) (mod, name string) {
	switch t := expr.(type) {
	case *ast.Ident:
		return currPath, t.Name
	case *ast.SelectorExpr:
		return getSelectorExprName(t, file)
	case *ast.StarExpr:
		return getExprName(file, currPath, t.X)
	default:
		panic(fmt.Sprintf("未处理的 ast.IndexExpr.Index 类型 %s", t))
	}
}

func buildTypeParams(list *ast.FieldList, tps []*typeParam) map[string]*typeParam {
	if list == nil || len(list.List) == 0 {
		return nil
	}

	m := make(map[string]*typeParam, len(list.List))
	for i, e := range list.List {
		m[e.Names[0].Name] = tps[i]
	}
	return m
}
