// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package enum 生成枚举的部分常用方法
package enum

import (
	"flag"
	"go/importer"
	"go/token"
	"go/types"
	"io"
	"io/fs"
	"slices"
	"strings"

	"github.com/issue9/cmdopt"
	"github.com/issue9/localeutil"
	"github.com/issue9/source/codegen"
	"github.com/issue9/web"
)

// 允许使用枚举的类型
//
// NOTE: 目前仅支持数值类型，如果是非数值类型，
// 那么在生成的 Parse 等方法中无法确定零值的表达方式。
var allowTypes = []string{
	"int", "int8", "int16", "int32", "int64",
	"uint", "uint8", "uint16", "uint32", "uint64",
	"float32", "float64",
}

var ErrNotAllowedType = web.NewLocaleError("not allowed enum type")

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

			ts := strings.Split(*t, ",")
			for i, name := range ts {
				ts[i] = strings.TrimSpace(name)
			}
			return dump(*h, *i, *o, ts)
		}
	})
}

func getValues(pkg *types.Package, types []string) (map[string][]string, error) {
	vals := make(map[string][]string, len(types))
	for _, t := range types {
		v, err := GetValue(pkg, t)
		if err != nil {
			return nil, err
		}
		vals[t] = v
	}

	return vals, nil
}

// GetValue 在 pkg 中查找类型名称为 t 的所有枚举值
func GetValue(pkg *types.Package, t string) ([]string, error) {
	typ, err := checkType(pkg, t)
	if err != nil {
		return nil, err
	}

	s := pkg.Scope()
	vals := make([]string, 0, 10)
	for _, v := range s.Names() {
		obj := s.Lookup(v)
		c, ok := obj.(*types.Const)
		if !ok {
			continue
		}
		if c.Type() == typ {
			vals = append(vals, c.Name())
		}
	}

	return vals, nil
}

// 检测类型是否存在于 f 且类型是允许使用的类型
func checkType(pkg *types.Package, t string) (types.Type, error) {
	obj := pkg.Scope().Lookup(t)
	if obj == nil {
		return nil, web.NewLocaleError("not found enum type %s", t)
	}

	tn, ok := obj.(*types.TypeName)
	if !ok {
		return nil, web.NewLocaleError("not found enum type %s", t)
	}

	if slices.Index(allowTypes, tn.Type().Underlying().String()) >= 0 {
		return tn.Type(), nil
	}

	return nil, ErrNotAllowedType
}

func dump(header, input, output string, ts []string) error {
	input = strings.TrimLeft(input, "./\\")
	fset := token.NewFileSet()
	imp := importer.ForCompiler(fset, "source", nil)
	pkg, err := imp.(types.ImporterFrom).ImportFrom(input, ".", 0)
	if err != nil {
		return err
	}

	vals, err := getValues(pkg, ts)
	if err != nil {
		return err
	}

	d := &data{
		FileHeader: header,
		Package:    pkg.Name(),
		Enums:      make([]*enum, 0, len(vals)),
	}

	for k, v := range vals {
		d.Enums = append(d.Enums, newEnum(k, v...))
	}

	return codegen.DumpFromTemplate(output, tpl, d, fs.ModePerm)
}
