// SPDX-License-Identifier: MIT

// Package enum 生成枚举的部分常用方法
package enum

import (
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"slices"
	"strings"

	"github.com/issue9/cmdopt"
	"github.com/issue9/errwrap"
	"github.com/issue9/localeutil"
	"github.com/issue9/source/codegen"
	"github.com/issue9/web"
)

// 允许使用枚举的类型
//
// NOTE: 目前仅支持数值类型，如果是非数据类型，
// 那么在生成的 Parse 等方法中无法确定零值的表达方式。
var allowTypes = []string{
	"int", "int8", "int16", "int32", "int64",
	"uint", "uint8", "uint16", "uint32", "uint64",
	"float32", "float64",
}

var errNotAllowedType = web.NewLocaleError("not allowed enum type")

const fileHeader = "当前文件由 web 生成，请勿手动编辑！"

const (
	title = web.StringPhrase("build enum file")
	usage = web.StringPhrase(`build enum file

flags:
{{flags}}`)

	outUsage   = web.StringPhrase("set output file")
	fhUsage    = web.StringPhrase("set file header")
	typeUsage  = web.StringPhrase("set the enum type")
	inputUsage = web.StringPhrase("set input file")
)

func Init(opt *cmdopt.CmdOpt, p *localeutil.Printer) {
	opt.New("enum", title.LocaleString(p), usage.LocaleString(p), func(fs *flag.FlagSet) cmdopt.DoFunc {
		o := fs.String("o", "", outUsage.LocaleString(p))
		h := fs.String("h", fileHeader, fhUsage.LocaleString(p))
		t := fs.String("t", "", typeUsage.LocaleString(p))
		i := fs.String("i", "", inputUsage.LocaleString(p))

		return func(io.Writer) error {
			if *i == "" {
				return web.NewLocaleError("no input file")
			}
			if *o == "" {
				return web.NewLocaleError("no output file")
			}
			if *t == "" {
				return web.NewLocaleError("type not set")
			}

			types := strings.Split(*t, ",")
			for i, name := range types {
				types[i] = strings.TrimSpace(name)
			}
			return dump(*h, *i, *o, types)
		}
	})
}

func getValues(f *ast.File, types []string) (map[string][]string, error) {
	vals := make(map[string][]string, len(types))
	for _, t := range types {
		v, err := getValue(f, t)
		if err != nil {
			return nil, err
		}
		vals[t] = v
	}

	return vals, nil
}

func getValue(f *ast.File, t string) ([]string, error) {
	if err := checkType(f, t); err != nil {
		return nil, err
	}

	vals := make([]string, 0, len(f.Decls))
	for _, d := range f.Decls {
		g, ok := d.(*ast.GenDecl)
		if !ok {
			continue
		}

		var isIota bool
		for _, s := range g.Specs {
			v, ok := s.(*ast.ValueSpec)
			if !ok {
				continue
			}

			if isIota {
				if v.Type == nil && len(v.Values) == 0 { // 未指定类型和值，则继承自上一条的类型
					vals = appendNames(vals, v.Names)
					continue
				}
				isIota = false

				// TODO(1): 如果需要支持其它类型，此处也需要作相应修改。
				if tt, ok := v.Type.(*ast.Ident); !ok || tt.Name != t {
					continue
				}

				vals = appendNames(vals, v.Names)
			} else {
				if v.Type == nil {
					continue
				}

				// 判断类型是否符合
				if tt, ok := v.Type.(*ast.Ident); !ok || tt.Name != t {
					continue
				}

				vals = appendNames(vals, v.Names)

				if vt, ok := v.Values[0].(*ast.Ident); ok {
					isIota = vt.Name == "iota"
				}
			}
		}
	}

	return vals, nil
}

func appendNames(val []string, v []*ast.Ident) []string {
	for _, n := range v {
		val = append(val, n.Name)
	}
	return val
}

// 检测类型是否存在于 f 且类型是允许使用的类型
func checkType(f *ast.File, t string) error {
	for _, d := range f.Decls {
		g, ok := d.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, s := range g.Specs {
			ts, ok := s.(*ast.TypeSpec)
			if !ok {
				continue
			}

			if ts.Name.Name == t {
				tt, ok := ts.Type.(*ast.Ident)

				// TODO(1): 只要底层类型在 allowTypes 的都应该放行
				if !ok || slices.Index(allowTypes, tt.Name) < 0 {
					return errNotAllowedType
				}
				return nil
			}
		}
	}

	return web.NewLocaleError("not found type %s", t)
}

func dump(header, input, output string, types []string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, input, nil, parser.AllErrors)
	if err != nil {
		return err
	}

	vals, err := getValues(f, types)
	if err != nil {
		return err
	}

	buf := &errwrap.Buffer{}

	buf.Printf("// %s \n\n", header).
		Printf("package %s \n\n", f.Name.Name).
		WString(`import (`).WByte('\n').
		WString(`"fmt"`).WByte('\n').
		WString(`"github.com/issue9/web"`).WString("\n").
		WString(`"github.com/issue9/web/locales"`).WByte('\n').
		WString(`)`).WByte('\n').WByte('\n')

	data := &Data{
		FileHeader: fileHeader,
		Package:    f.Name.Name,
		Types:      make([]*Type, 0, len(vals)),
	}

	for k, v := range vals {
		data.Types = append(data.Types, NewType(k, v...))
	}

	return codegen.DumpFromTemplate(output, tpl, data, fs.ModePerm)
}
