// SPDX-License-Identifier: MIT

//go:build ignore

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/issue9/errwrap"
)

const (
	filename = "CONFIG.html"
	objName  = "configOf"
)

type table struct {
	name string
	w    *errwrap.Writer
}

var primitiveTypes = []string{
	"int", "int8", "int16", "int32", "int64",
	"uint", "uint8", "uint16", "uint32", "uint64",
	"string", "bool",

	// 无须处理的自定义类型
	"duration",
	"logs.Level",
}

func main() {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, "./", nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	var p *ast.Package
	for name, pkg := range pkgs {
		if name == "main" || strings.HasSuffix(name, "_test") {
			continue
		}

		p = pkg
		break
	}

	if p == nil {
		panic("未找到正确的包")
	}

	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := &errwrap.Writer{Writer: f}

	w.WString(`<!DOCTYPE html>
<html lang="zh-CN">
	<head>
		<title>config</title>
		<meta charset="utf-8" />
		<style>
		:root {
			--color: black;
			--bg: white;
		}
		@media (prefers-color-scheme: dark) {
			:root {
				--color: white;
				--bg: black;
			}
		}
		table {
			width: 100%;
			border-collapse: collapse;
			border: 1px solid var(--color);
			text-align: left;
		}
		th {
			text-align: left;
		}
		tr {
			border-bottom: 1px solid var(--color);
		}

		body {
			color: var(--color);
			background: var(--bg);
		}
		</style>
	</head>
	<body>`)

	w.WString(`<h1>config</h1>`).WString(`<article>
	这是 LoadOptions 用到的配置项。
	</article>`)

	parse(w, objName, p)

	w.WString(`
	<p><a href="https://github.com/caixw">作者</a></p>
	<p><a href="https://github.com/issue9/web">代码仓库</a></p>
	</body>
</html>`)
}

func parse(w *errwrap.Writer, outputObj string, pkg *ast.Package) {
	waitList := make([]string, 0, 10)

	for _, f := range pkg.Files {
		for _, decl := range f.Decls {
			d, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			for _, spec := range d.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				if ts.Name.Name == outputObj {
					waitList = append(waitList, parseObject(w, ts)...)
				}
			}
		}
	}

	for _, obj := range waitList {
		parse(w, obj, pkg)
	}
}

func parseObject(w *errwrap.Writer, obj *ast.TypeSpec) []string {
	s, ok := obj.Type.(*ast.StructType)
	if !ok {
		panic(fmt.Sprintf("%v 不能转换成 ast.StructType", obj))
	}

	waitList := make([]string, 0, 10)

	t := newTable(w, obj.Name.Name, obj.Doc.Text())
	defer t.end()

	for _, f := range s.Fields.List {
		name := f.Names[0].Name
		if !ast.IsExported(name) || name == "XMLName" {
			continue
		}

		var xml, json, yaml string
		if f.Tag != nil {
			st := reflect.StructTag(strings.Trim(f.Tag.Value, "`"))
			xml = getName(name, st.Get("xml"))
			json = getName(name, st.Get("json"))
			yaml = getName(name, st.Get("yaml"))
		}

		var fieldTypeName string
		switch ft := f.Type.(type) {
		case *ast.StarExpr:
			fieldTypeName = ft.X.(*ast.Ident).Name
		case *ast.Ident:
			fieldTypeName = ft.Name
		case *ast.SelectorExpr:
			fieldTypeName = ft.X.(*ast.Ident).Name + "." + ft.Sel.Name
		case *ast.ArrayType:
			switch elt := ft.Elt.(type) {
			case *ast.StarExpr:
				fieldTypeName = elt.X.(*ast.Ident).Name
			case *ast.Ident:
				fieldTypeName = elt.Name
			case *ast.SelectorExpr:
				fieldTypeName = elt.X.(*ast.Ident).Name + "." + elt.Sel.Name
			default:
				panic(fmt.Sprintf("字段 %s 无法转换成 *ast.Ident", f.Names[0].Name))
			}
		default:
			panic(fmt.Sprintf("字段 %s 无法转换成 *ast.Ident", f.Names[0].Name))
		}

		if slices.Index(primitiveTypes, fieldTypeName) < 0 {
			waitList = append(waitList, fieldTypeName)
			fieldTypeName = `<a href="#` + fieldTypeName + `">` + fieldTypeName + "</a>"
		}

		t.write(xml, json, yaml, fieldTypeName, f.Doc.Text())
	}

	return waitList
}

func getName(name string, tag string) string {
	if tag == "-" {
		return "-"
	}

	if tag == "" {
		return name
	}
	return tag
}

func newTable(w *errwrap.Writer, name, desc string) *table {
	w.Printf(`<h2 id="%s">%s</h2>`, name, name).WByte('\n').
		WString("<article>").WString(desc).WString("</article>").WByte('\n')

	w.WString(`<table>
	<thead><tr><th>XML</th><th>JSON</th><th>YAML</th><th>类型</th><th>说明</th></tr></thead>
	<tbody>`)

	return &table{
		name: name,
		w:    w,
	}
}

var brReplacer = strings.NewReplacer("\n", "<br />")

func (t *table) write(xml, json, yaml, typ, desc string) {
	desc = brReplacer.Replace(desc)
	t.w.Printf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`, xml, json, yaml, typ, desc)
}

func (t *table) end() {
	t.w.WString("</tbody></table>")
}
