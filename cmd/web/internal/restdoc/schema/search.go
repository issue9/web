// SPDX-License-Identifier: MIT

package schema

import (
	"go/ast"
	"go/token"
	"path"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/localeutil"
	"github.com/issue9/query/v3"
	"github.com/issue9/web"

	"github.com/issue9/web/cmd/web/internal/restdoc/pkg"
	"github.com/issue9/web/cmd/web/internal/restdoc/utils"
)

var refReplacer = strings.NewReplacer("/", ".")

type SearchFunc func(string) *pkg.Package

// currPath 当前包的导出路径；
// typeName 表示需要查找的类型名，非内置类型且不带路径信息，则将 currPath 作为路径信息。
// q 是否用于查询参数
//
// 可能返回的错误值为 *Error
func (f SearchFunc) New(t *openapi3.T, currPath, typeName string, q bool) (*openapi3.SchemaRef, error) {
	var isArray bool
	if strings.HasPrefix(typeName, "[]") {
		typeName = typeName[2:]
		isArray = true
	}

	tag := "json"
	if q {
		tag = query.Tag
	}

	return f.fromName(t, currPath, typeName, tag, isArray)
}

// 根据类型名生成 schema 对象
//
// 参数参考 [SearchFunc.New]
func (f SearchFunc) fromName(t *openapi3.T, currPath, typeName, tag string, isArray bool) (*openapi3.SchemaRef, error) {
	switch typeName { // 基本类型
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return array(openapi3.NewSchemaRef("", openapi3.NewIntegerSchema()), isArray), nil
	case "float32", "float64":
		return array(openapi3.NewSchemaRef("", openapi3.NewFloat64Schema()), isArray), nil
	case "bool":
		return array(openapi3.NewSchemaRef("", openapi3.NewBoolSchema()), isArray), nil
	case "string":
		return array(openapi3.NewSchemaRef("", openapi3.NewStringSchema()), isArray), nil
	case "map":
		return array(openapi3.NewSchemaRef("", openapi3.NewObjectSchema()), isArray), nil
	}

	modPath := currPath
	modName := typeName
	if index := strings.LastIndexByte(typeName, '.'); index > 0 { // 全局的路径
		modPath = typeName[:index]
		modName = typeName[index+1:]
	} else {
		typeName = currPath + "." + typeName
	}
	if modPath == "" {
		return nil, localeutil.Error("not found %s", typeName) // 行数未变化，直接返回错误。
	}

	ref := refReplacer.Replace(typeName)

	if schemaRef, found := t.Components.Schemas[ref]; found { // 查找是否已经存在于 components/schemes
		sr := openapi3.NewSchemaRef(ref, schemaRef.Value)
		addRefPrefix(sr)
		return array(sr, isArray), nil
	}

	pkg := f(modPath)
	if pkg == nil {
		return nil, localeutil.Error("not found %s", modPath) // 行数未变化，直接返回错误。
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

	schemaRef, err := f.fromTypeSpec(t, file, currPath, ref, tag, spec)
	if err != nil {
		return nil, err
	}

	if schemaRef.Ref != "" && tag != query.Tag { // 查询参数不保存整个对象
		t.Components.Schemas[schemaRef.Ref] = openapi3.NewSchemaRef("", schemaRef.Value)
		addRefPrefix(schemaRef)
	}
	return array(schemaRef, isArray), nil
}

// 将 ast.TypeSpec 转换成 openapi3.Schema
//
// typeName 仅用于生成 SchemaRef.Ref 值，需要完整路径。
func (f SearchFunc) fromTypeSpec(t *openapi3.T, file *ast.File, currPath, ref, tag string, s *ast.TypeSpec) (*openapi3.SchemaRef, error) {
	desc, enums := parseTypeDoc(s)
	if desc == "" && s.Comment != nil {
		desc = s.Comment.Text()
	}

	switch ts := s.Type.(type) {
	case *ast.Ident: // type x = int 或是 type x int
		schemaRef, err := f.fromName(t, currPath, ts.Name, tag, false)
		if err != nil {
			return nil, newError(s.Pos(), err)
		}
		schemaRef.Value.Description = desc
		schemaRef.Value.Enum = enums
		schemaRef.Ref = ref
		return schemaRef, nil
	case *ast.SelectorExpr: // type x = json.Decoder 或是 type x json.Decoder 引用外部对象
		name := getSelectorExprTypeName(ts, file)
		schemaRef, err := f.fromName(t, currPath, name, tag, false)
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

		if err := f.addFields(t, file, schema, currPath, tag, ts.Fields.List); err != nil {
			return nil, err
		}

		return openapi3.NewSchemaRef(ref, schema), nil
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

// 将 list 中的所有字段解析到 schema
//
// 字段名如果存在 json 时，取 json 名称，否则直接采用字段名，xml 仅采用了 attr 和 parent>child 两种格式。
func (f SearchFunc) addFields(t *openapi3.T, file *ast.File, s *openapi3.Schema, modPath, tagName string, list []*ast.Field) error {
LOOP:
	for _, field := range list {
		if len(field.Names) == 0 { // 嵌套对象
			ref, err := f.fromExpr(t, file, modPath, tagName, field.Type)
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

		item, err := f.fromExpr(t, file, modPath, tagName, field.Type)
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

// 将 s 中的内容转换到 schema 上
func (f SearchFunc) fromExpr(t *openapi3.T, file *ast.File, currPath, tag string, e ast.Expr) (*openapi3.SchemaRef, error) {
	switch expr := e.(type) {
	case *ast.ArrayType:
		schema, err := f.fromExpr(t, file, currPath, tag, expr.Elt)
		if err != nil {
			return nil, err
		}
		return array(schema, true), nil
	case *ast.MapType: // NOTE: map 无法指定字段名
		return openapi3.NewSchemaRef("", openapi3.NewObjectSchema()), nil
	case *ast.Ident:
		ref, err := f.fromName(t, currPath, expr.Name, tag, false)
		if err != nil {
			return nil, newError(e.Pos(), err)
		}
		return ref, nil
	case *ast.StarExpr: // 指针
		return f.fromExpr(t, file, currPath, tag, expr.X)
	case *ast.SelectorExpr:
		name := getSelectorExprTypeName(expr, file)
		ref, err := f.fromName(t, currPath, name, tag, false)
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

func getSelectorExprTypeName(expr *ast.SelectorExpr, file *ast.File) string {
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
			pkgName = p
			break
		}
	}
	return pkgName + "." + expr.Sel.Name
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
func array(ref *openapi3.SchemaRef, isArray bool) *openapi3.SchemaRef {
	if !isArray {
		return ref
	}

	s := openapi3.NewArraySchema()
	s.Items = ref
	return openapi3.NewSchemaRef("", s)
}

// 将从 components/schemas 中获取的对象进行二次包装
func wrap(ref *openapi3.SchemaRef, desc string, xml *openapi3.XML, nullable bool) *openapi3.SchemaRef {
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
		ref = openapi3.NewSchemaRef("", s)
	}
	return ref
}

const refPrefix = "#/components/schemas/"

func addRefPrefix(ref *openapi3.SchemaRef) {
	if ref.Ref != "" && !strings.HasPrefix(ref.Ref, refPrefix) {
		ref.Ref = refPrefix + ref.Ref
	}
}
