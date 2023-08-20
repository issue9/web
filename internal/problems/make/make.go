// SPDX-License-Identifier: MIT

// Package make 提供一些通用的生成工具
package make

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
)

const FileHeader = "// 此文件由工具产生，请勿手动修改！\n\n"

func ID(p Pair) string { return "Problem" + p.Name[6:] }

// GetStatuses 从 net/http/status.go 获取所有的状态码
func GetStatuses() ([]Pair, error) {
	path := filepath.Join(build.Default.GOROOT, "src", "net", "http", "status.go")
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	pairs := make([]Pair, 0, 100)

LOOP:
	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)

			if !ok {
				continue LOOP
			}

			name := vs.Names[0].Name
			if name != "_" && !strings.HasPrefix(name, "Status") { // name 格式不正确
				continue LOOP
			}

			val, ok := vs.Values[0].(*ast.BasicLit)
			if !ok || val.Kind != token.INT { // value 不正确
				continue LOOP
			}

			v, err := strconv.Atoi(val.Value)
			if err != nil {
				return nil, err
			}

			if v < 400 {
				continue
			} // 400 以下的不需要

			pairs = append(pairs, Pair{Name: name, Value: v})
		}
	}

	return pairs, nil
}

type Pair struct {
	Name  string
	Value int
}