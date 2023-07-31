// SPDX-License-Identifier: MIT

package schema

import (
	"fmt"
	"go/ast"
	"go/token"
	"path"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/localeutil"
	"github.com/issue9/query/v3"
	"github.com/issue9/sliceutil"
	"github.com/issue9/web"

	"github.com/issue9/web/cmd/web/internal/restdoc/pkg"
	"github.com/issue9/web/cmd/web/internal/restdoc/utils"
)

type SearchFunc func(string) *pkg.Package

// currPath 当前包的导出路径；
// typeName 表示需要查找的类型名，非内置类型且不带路径信息，则将 currPath 作为路径信息。
// q 是否用于查询参数
//
// 可能返回的错误值为 *Error
func (f SearchFunc) New(t *openapi3.T, currPath, typeName string, q bool) (*Ref, error) {
	var isArray bool
	if strings.HasPrefix(typeName, "[]") {
		typeName = typeName[2:]
		isArray = true
	}

	tag := "json"
	if q {
		tag = query.Tag
	}

	return f.fromName(t, currPath, typeName, tag, isArray, nil)
}

// 根据类型名生成 schema 对象
//
// tpRefs 泛型参数对应的 Ref，非泛型则为空；
// 其它参数参考 [SearchFunc.New]
func (f SearchFunc) fromName(t *openapi3.T, currPath, typeName, tag string, isArray bool, tpRefs []*Ref) (*Ref, error) {
	switch typeName { // 基本类型
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return array(NewRef("", openapi3.NewIntegerSchema()), isArray), nil
	case "float32", "float64":
		return array(NewRef("", openapi3.NewFloat64Schema()), isArray), nil
	case "bool":
		return array(NewRef("", openapi3.NewBoolSchema()), isArray), nil
	case "string":
		return array(NewRef("", openapi3.NewStringSchema()), isArray), nil
	case "map":
		return array(NewRef("", openapi3.NewObjectSchema()), isArray), nil
	}

	modName := typeName
	if index := strings.LastIndexByte(typeName, '.'); index > 0 { // 全局的路径
		currPath = typeName[:index]
		modName = typeName[index+1:]
	} else {
		typeName = currPath + "." + typeName
	}
	if currPath == "" {
		return nil, localeutil.Error("not found %s", typeName) // 行数未变化，直接返回错误。
	}

	ref := refReplacer.Replace(typeName)

	if schemaRef, found := t.Components.Schemas[ref]; found { // 查找是否已经存在于 components/schemes
		sr := NewRef(ref, schemaRef.Value)
		addRefPrefix(sr)
		return array(sr, isArray), nil
	}

	pkg := f(currPath)
	if pkg == nil {
		return nil, localeutil.Error("not found %s", currPath) // 行数未变化，直接返回错误。
	}

	var spec *ast.TypeSpec
	var file *ast.File
LOOP:
	for _, f := range pkg.Files {
		for _, d := range f.Decls {
			gen, ok := d.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}

			for _, s := range gen.Specs {
				if spec, ok = s.(*ast.TypeSpec); ok && spec.Name.Name == modName {
					file = f
					break LOOP // 找到了，就退到最外层。
				}
			}
		}
	}

	if spec == nil || file == nil {
		return nil, localeutil.Error("not found %s", typeName)
	}

	schemaRef, err := f.fromTypeSpec(t, file, currPath, ref, tag, spec, tpRefs)
	if err != nil {
		return nil, err
	}

	if schemaRef.Ref != "" &&
		tag != query.Tag && // 查询参数不保存整个对象
		spec.TypeParams == nil { // 泛型不保存
		t.Components.Schemas[schemaRef.Ref] = NewRef("", schemaRef.Value)
		addRefPrefix(schemaRef)
	}
	return array(schemaRef, isArray), nil
}

// 将 ast.TypeSpec 转换成 openapi3.SchemaRef
//
// ref 仅用于生成 SchemaRef.Ref 值，需要完整路径。
func (f SearchFunc) fromTypeSpec(t *openapi3.T, file *ast.File, currPath, ref, tag string, s *ast.TypeSpec, tpRefs []*Ref) (*Ref, error) {
	desc, enums := parseTypeDoc(s)
	if desc == "" && s.Comment != nil {
		desc = s.Comment.Text()
	}

	switch ts := s.Type.(type) {
	case *ast.Ident: // type x = int 或是 type x int
		schemaRef, err := f.fromName(t, currPath, ts.Name, tag, false, nil)
		if err != nil {
			return nil, newError(s.Pos(), err)
		}
		schemaRef.Value.Description = desc
		schemaRef.Value.Enum = enums
		schemaRef.Ref = ref
		return schemaRef, nil
	case *ast.IndexExpr: // type x = G[int]
		return f.fromIndexExpr(t, file, currPath, tag, ts)
	case *ast.IndexListExpr: // type x = G[int, float]
		return f.fromIndexListExpr(t, file, currPath, ref, tag, ts)
	case *ast.SelectorExpr: // type x = json.Decoder 或是 type x json.Decoder 引用外部对象
		mod, name := getSelectorExprName(ts, file)
		schemaRef, err := f.fromName(t, mod, name, tag, false, nil)
		if err != nil {
			return nil, newError(s.Pos(), err)
		}
		schemaRef.Value.Description = desc
		schemaRef.Value.Enum = enums
		return schemaRef, nil
	case *ast.StructType:
		schema := openapi3.NewObjectSchema()
		schema.Description = desc
		schema.Enum = enums

		if err := f.addFields(t, file, schema, currPath, tag, ts.Fields.List, s.TypeParams, tpRefs); err != nil {
			return nil, err
		}

		return NewRef(ref, schema), nil
	default:
		msg := web.Phrase("%s can not convert to ast.StructType", s.Type)
		return nil, newError(s.Pos(), msg)
	}
}

func parseTypeDoc(s *ast.TypeSpec) (desc string, enums []any) {
	if s.Doc == nil {
		return "", nil
	}
	text := s.Doc.Text()
	if text == "" {
		return "", nil
	}

	// @enum e1 e2 e3
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if tag, suffix := utils.CutTag(line); tag == "@enum" {
			for _, word := range strings.Fields(suffix) {
				enums = append(enums, word)
			}
		}
	}

	return text, enums
}

func (f SearchFunc) fromIndexExpr(t *openapi3.T, file *ast.File, currPath, tag string, idx *ast.IndexExpr) (*Ref, error) {
	mod, idxName := getExprName(file, currPath, idx.Index)
	idxRef, err := f.fromName(t, mod, idxName, tag, false, nil)
	if err != nil {
		return nil, err
	}

	mod, name := getExprName(file, currPath, idx.X)
	return f.fromName(t, mod, name, tag, false, []*Ref{idxRef})
}

func (f SearchFunc) fromIndexListExpr(t *openapi3.T, file *ast.File, currPath, ref, tag string, idx *ast.IndexListExpr) (*Ref, error) {
	indexes := make([]*Ref, 0, len(idx.Indices))

	for _, i := range idx.Indices {
		mod, idxName := getExprName(file, currPath, i)
		idxRef, err := f.fromName(t, mod, idxName, tag, false, nil)
		if err != nil {
			return nil, err
		}
		indexes = append(indexes, idxRef)
	}

	mod, name := getExprName(file, currPath, idx.X)
	return f.fromName(t, mod, name, tag, false, indexes)
}

// 将 fields 中的所有字段解析到 schema
//
// 字段名如果存在 json 时，取 json 名称，否则直接采用字段名，xml 仅采用了 attr 和 parent>child 两种格式。
func (f SearchFunc) addFields(t *openapi3.T, file *ast.File, s *openapi3.Schema, modPath, tagName string, fields []*ast.Field, tp *ast.FieldList, tpRefs []*Ref) error {
LOOP:
	for _, field := range fields {
		if len(field.Names) == 0 { // 嵌套对象
			ref, err := f.fromExpr(t, file, modPath, tagName, field.Type, nil, nil)
			if err != nil {
				return err
			}

			for k, v := range ref.Value.Properties {
				s.WithPropertyRef(k, v)
			}
			continue
		}

		name, nullable, xml := parseTag(field, tagName)
		if name == "-" {
			continue LOOP
		}

		item, err := f.fromExpr(t, file, modPath, tagName, field.Type, tp, tpRefs)
		if err != nil {
			return err
		}

		var desc string
		if field.Doc != nil {
			desc = field.Doc.Text()
		}
		if desc == "" && field.Comment != nil {
			desc = field.Comment.Text()
		}

		s.WithPropertyRef(name, wrap(item, desc, xml, nullable))
	}

	return nil
}

// 将 ast.Expr 中的内容转换到 schema 上
func (f SearchFunc) fromExpr(t *openapi3.T, file *ast.File, currPath, tag string, e ast.Expr, tp *ast.FieldList, tpRefs []*Ref) (*Ref, error) {
	switch expr := e.(type) {
	case *ast.ArrayType:
		schema, err := f.fromExpr(t, file, currPath, tag, expr.Elt, tp, tpRefs)
		if err != nil {
			return nil, err
		}
		return array(schema, true), nil
	case *ast.MapType: // NOTE: map 无法指定字段名
		return NewRef("", openapi3.NewObjectSchema()), nil
	case *ast.Ident:
		if len(tpRefs) > 0 { // 这是泛型类型
			if index := sliceutil.Index(tp.List, func(item *ast.Field, _ int) bool { return item.Names[0].Name == expr.Name }); index >= 0 {
				return tpRefs[index], nil
			} else {
				panic(fmt.Sprintf("无法为泛型的类型参数 %s 找到对应的类型", expr.Name))
			}
		}
		return f.fromName(t, currPath, expr.Name, tag, false, nil)
	case *ast.StarExpr: // 指针
		return f.fromExpr(t, file, currPath, tag, expr.X, tp, tpRefs)
	case *ast.SelectorExpr:
		mod, name := getSelectorExprName(expr, file)
		ref, err := f.fromName(t, mod, name, tag, false, tpRefs)
		if err != nil {
			if _, ok := err.(*Error); ok {
				return nil, err
			}
			return nil, newError(e.Pos(), err)
		}
		return ref, nil
	//case *ast.InterfaceType: // 无法处理此类型
	default:
		return nil, newError(e.Pos(), web.Phrase("unsupported ast expr %s", expr))
	}
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

func getSelectorExprName(expr *ast.SelectorExpr, file *ast.File) (mod, name string) {
	pkgName := expr.X.(*ast.Ident).Name
	for _, d := range file.Imports {
		p := strings.Trim(d.Path.Value, "\"")

		var name string
		if d.Name != nil {
			name = d.Name.Name
		} else {
			name = path.Base(p)
		}

		if name == pkgName {
			return p, expr.Sel.Name
		}
	}
	panic(fmt.Sprintf("无法找到 %s 的导入名称", pkgName))
}

func parseTag(field *ast.Field, tagName string) (name string, nullable bool, xml *openapi3.XML) {
	name = field.Names[0].Name
	if !token.IsExported(name) { // 不能导出的字段自动忽略
		return "-", false, nil
	}

	if field.Tag != nil {
		structTag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
		tag := structTag.Get(tagName)
		if tag == "-" { // 忽略此字段
			return "-", false, nil
		}

		if tag != "" {
			words := strings.Split(tag, ",")
			name = words[0]
			if len(words) > 1 && words[1] == "omitempty" {
				nullable = true
			}
		}

		if tagName != query.Tag { // 非查询参数对象，需要处理 XML 的特殊情况
			tag := structTag.Get("xml")
			if tag != "" && tag != "-" {
				words := strings.Split(tag, ",")
				switch len(words) {
				case 1:
					if strings.IndexByte(words[0], '>') > 0 {
						xml = &openapi3.XML{Wrapped: true}
					}
				case 2:
					wrap := strings.IndexByte(words[0], '>') > 0
					attr := words[1] == "attr"
					if wrap || attr {
						xml = &openapi3.XML{Wrapped: wrap, Attribute: attr}
					}
				}
			}
		}
	}
	return
}

// 根据 isArray 将 ref 包装成相应的对象
func array(ref *Ref, isArray bool) *Ref {
	if !isArray {
		return ref
	}

	s := openapi3.NewArraySchema()
	s.Items = ref
	return NewRef("", s)
}

// 将从 components/schemas 中获取的对象进行二次包装
func wrap(ref *Ref, desc string, xml *openapi3.XML, nullable bool) *Ref {
	if ref == nil {
		return ref
	}

	if ref.Value.Nullable != nullable ||
		ref.Value.XML != xml ||
		(desc != "" && ref.Value.Description != desc) {
		s := openapi3.NewSchema()
		s.AllOf = openapi3.SchemaRefs{ref}
		s.Nullable = nullable
		s.XML = xml
		if desc != "" {
			s.Description = desc
		}
		ref = NewRef("", s)
	}
	return ref
}
