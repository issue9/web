// SPDX-License-Identifier: MIT

package htmldoc

import (
	"fmt"
	"go/ast"
	"go/doc/comment"
	"go/parser"
	"go/token"
	"html/template"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/issue9/localeutil"
	"github.com/issue9/web"

	"github.com/issue9/web/cmd/web/locales"
)

type data struct {
	Title  string // 标题，页面标题和文档的 h1
	Desc   template.HTML
	Style  template.CSS // stylesheet 样式表
	Lang   string       // 页面语言
	Header template.HTML
	Footer template.HTML

	TypeLocale string // 表格中 type 的翻译项
	DescLocale string // 表格中 desc 的翻译项
	Objects    []*object
}

type object struct {
	Title string
	Desc  template.HTML
	Items []*item
}

type item struct {
	XML, JSON, YAML string
	Type            template.HTML
	Desc            template.HTML
}

var basicTypes = []string{
	"int", "int8", "int16", "int32", "int64",
	"uint", "uint8", "uint16", "uint32", "uint64",
	"string", "bool",
}

// 导出为 html 文档
//
// dir 为源码目录；
// objName 表示从源码目录中需要提取的对象；
// output 输出的 html 文档路径；
// lang 输出的文档语言，被应用在 html 的 lang 属性上；
// title 文档的标题；
// desc 文档的描述，可以是 HTML 格式；
// style 样式表；
func export(dir, objName, output, lang, title, desc, header, footer, style string) error {
	p, err := locales.NewPrinter(lang)
	if err != nil {
		return err
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	var files map[string]*ast.File
	for name, pkg := range pkgs {
		if name != "main" && !strings.HasSuffix(name, "_test") { // 过滤测试包和 make_* 的包
			files = pkg.Files
			break
		}
	}

	if files == nil {
		return web.NewLocaleError("not found source in %s", dir)
	}

	if style == defaultStyleValue {
		style = defaultStyle
	}

	d := &data{
		Title:  title,
		Desc:   template.HTML(desc),
		Style:  template.CSS(style),
		Lang:   lang,
		Header: template.HTML(header),
		Footer: template.HTML(footer),

		TypeLocale: web.StringPhrase("type").LocaleString(p),
		DescLocale: web.StringPhrase("description").LocaleString(p),
		Objects:    make([]*object, 0, 100),
	}
	d.parse(p, objName, files)

	t, err := template.New("htmldoc").Parse(tpl)
	if err != nil {
		return err
	}

	f, err := os.Create(output)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.Execute(f, d)
}

func (d *data) parse(p *localeutil.Printer, outputObj string, files map[string]*ast.File) {
	waitList := make([]string, 0, 10)

	var found bool
	for _, f := range files {
		for _, decl := range f.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			for _, spec := range gen.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				if ts.Name.Name == outputObj {
					found = true
					if len(gen.Specs) == 1 {
						ts.Doc = gen.Doc
					}
					waitList = append(waitList, d.parseObject(ts)...)
				}
			}
		}
	}

	if !found {
		d.append(&object{Title: outputObj, Desc: template.HTML(web.Phrase("not found type doc").LocaleString(p))})
	}

	for _, obj := range waitList {
		d.parse(p, obj, files)
	}
}

func (d *data) append(o *object) {
	if slices.IndexFunc(d.Objects, func(obj *object) bool { return obj.Title == o.Title }) >= 0 {
		return
	}
	d.Objects = append(d.Objects, o)
}

func (d *data) parseObject(obj *ast.TypeSpec) []string {
	s, ok := obj.Type.(*ast.StructType)
	if !ok {
		d.append(&object{
			Title: obj.Name.Name,
			Desc:  comment2HTML(obj.Doc, obj.Comment),
		})
		return nil
	}

	waitList := make([]string, 0, 10)

	o := &object{
		Title: obj.Name.Name,
		Desc:  comment2HTML(obj.Doc, obj.Comment),
		Items: make([]*item, 0, len(s.Fields.List)),
	}

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

		if slices.Index(basicTypes, fieldTypeName) < 0 {
			waitList = append(waitList, fieldTypeName)
			fieldTypeName = `<a href="#` + fieldTypeName + `">` + fieldTypeName + "</a>"
		}

		o.Items = append(o.Items, &item{
			XML:  xml,
			JSON: json,
			YAML: yaml,
			Type: template.HTML(fieldTypeName),
			Desc: comment2HTML(f.Doc, f.Comment),
		})
	}

	d.append(o)

	return waitList
}

// name 为默认名称
// tag 为 struct tag 的值
func getName(name, tag string) string {
	if tag == "-" {
		return "-"
	}

	if tag == "" {
		return name
	}
	return tag
}

var (
	cPrinter comment.Printer
	cParser  comment.Parser
)

func comment2HTML(doc, c *ast.CommentGroup) template.HTML {
	if doc == nil {
		doc = c
	}
	return template.HTML(string(cPrinter.HTML(cParser.Parse(doc.Text()))))
}
