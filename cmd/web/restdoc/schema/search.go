// SPDX-License-Identifier: MIT

package schema

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"path"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/query/v3"
	"github.com/issue9/web"

	"github.com/issue9/web/cmd/web/restdoc/openapi"
	"github.com/issue9/web/cmd/web/restdoc/pkg"
)

type SearchFunc func(string) *pkg.Package

// New 根据类型名称 typePath 生成 SchemaRef 对象
//
// currPath 当前包的导出路径；
// typePath 表示需要查找的类型名，非内置类型且不带路径信息，则将 currPath 作为路径信息；
// typePath 可以包含类型参数，比如 G[int]。
// 如果 typePath 以 #components/schemas 开头，则从 t.Components.Schemas 下查找。
// q 是否用于查询参数
//
// 可能返回的错误值为 *Error
func (f SearchFunc) New(t *openapi.OpenAPI, currPath, typePath string, q bool) (*Ref, error) {
	if strings.HasPrefix(typePath, refPrefix) {
		if ref, found := t.GetSchema(strings.TrimPrefix(typePath, refPrefix)); found {
			return ref, nil
		} else {
			return nil, web.NewLocaleError("not found schema ref %s", typePath)
		}
	}

	var isArray bool
	if strings.HasPrefix(typePath, "[]") {
		typePath = typePath[2:]
		isArray = true
	}

	tag := "json"
	if q {
		tag = query.Tag
	}

	return f.fromName(t, currPath, typePath, tag, isArray)
}

// 根据类型名生成 schema 对象
func (f SearchFunc) fromName(t *openapi.OpenAPI, currPath, typePath, tag string, isArray bool) (*Ref, error) {
	if r, found := getPrimitiveType(typePath, isArray); found {
		return r, nil
	}

	structPath := currPath
	structName := typePath
	pp := typePath
	if bIndex := strings.IndexByte(typePath, '['); bIndex > 0 {
		pp = typePath[:bIndex]
	}
	if index := strings.LastIndexByte(pp, '.'); index > 0 { // 全局的路径
		structPath = pp[:index]
		structName = typePath[index+1:]
	} else {
		typePath = currPath + "." + typePath
	}
	ref := refReplacer.Replace(typePath)

	if schemaRef, found := t.GetSchema(ref); found { // 查找是否已经存在于 components/schemes
		return array(NewRef(addRefPrefix(ref), schemaRef.Value), isArray), nil
	}

	var tpRefs []*typeParam // 如果是范型，拿到范型的参数。
	if index := strings.LastIndexByte(structName, '['); index > 0 && structName[len(structName)-1] == ']' {
		tps := strings.Split(structName[index+1:len(structName)-1], ",")
		structName = structName[:index] // 范型，去掉类型参数部分

		tpRefs = make([]*typeParam, 0, len(tps))
		for _, i := range tps {
			name := strings.TrimSpace(i)
			idxRef, err := f.fromName(t, currPath, name, tag, false)
			if err != nil {
				return nil, err
			}
			tpRefs = append(tpRefs, &typeParam{ref: idxRef, name: name})
		}
	}

	file, spec, err := f.findTypeSpec(structPath, structName)
	if err != nil {
		return nil, err
	}

	if spec.TypeParams != nil && len(tpRefs) == 0 {
		return nil, web.NewLocaleError("unsupported generics type %s", typePath)
	}

	schemaRef, err := f.fromTypeSpec(t, structPath, tag, file, spec, tpRefs)
	if err != nil {
		return nil, err
	}

	if tag != query.Tag || schemaRef.Value.Type != openapi3.TypeObject { // 查询参数不保存整个对象
		t.AddSchema(ref, schemaRef.Value)
	}

	schemaRef.Ref = addRefPrefix(ref)
	return array(schemaRef, isArray), nil
}

func (f SearchFunc) findTypeSpec(structPath, structName string) (file *ast.File, spec *ast.TypeSpec, err error) {
	p := f(structPath)
	if p == nil {
		return nil, nil, web.NewLocaleError("not found module %s", structPath)
	}

	for _, file = range p.Files {
		for _, d := range file.Decls {
			gen, ok := d.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}

			for _, s := range gen.Specs {
				if spec, ok = s.(*ast.TypeSpec); ok && spec.Name.Name == structName {
					if len(gen.Specs) == 1 { // 不在 type() 内声明的类型，而是 type x struct{} 的单个声明。
						spec.Doc = gen.Doc
					}
					return
				}
			}
		}
	}

	return nil, nil, web.NewLocaleError("not found struct %s", structPath+"."+structName)
}

// 将 ast.TypeSpec 转换成 openapi3.SchemaRef
func (f SearchFunc) fromTypeSpec(t *openapi.OpenAPI, currPath, tag string, file *ast.File, s *ast.TypeSpec, tpRefs []*typeParam) (*Ref, error) {
	title, desc, typ, enums := parseTypeDoc(s)

	if typ != "" { // 自定义了类型
		s := openapi3.NewSchema().WithEnum(enums...)
		s.Type = typ
		s.Title = title
		s.Description = desc
		return NewRef("", s), nil
	}

	switch ts := s.Type.(type) {
	case *ast.Ident: // type x = int 或是 type x int
		schemaRef, err := f.fromName(t, currPath, ts.Name, tag, false)
		if err != nil {
			return nil, newError(s.Pos(), err)
		}
		if title != "" || desc != "" || len(enums) > 0 && schemaRef.Ref != "" {
			s := openapi3.NewSchema().WithEnum(enums...)
			s.AllOf = openapi3.SchemaRefs{schemaRef}
			s.Title = title
			s.Description = desc
			return NewRef("", s), nil
		}
		return schemaRef, nil
	case *ast.SelectorExpr: // type x = json.Decoder 或是 type x json.Decoder 引用外部对象
		mod, name := getSelectorExprName(ts, file)
		schemaRef, err := f.fromName(t, currPath, mod+"."+name, tag, false)
		if err != nil {
			return nil, newError(s.Pos(), err)
		}
		if title != "" || desc != "" || len(enums) > 0 && schemaRef.Ref != "" {
			s := openapi3.NewSchema().WithEnum(enums...)
			s.AllOf = openapi3.SchemaRefs{schemaRef}
			s.Title = title
			s.Description = desc
			return NewRef("", s), nil
		}
		return schemaRef, nil
	case *ast.ArrayType: // type x = []int , type x[T any] = []T
		tps := buildTypeParams(s.TypeParams, tpRefs)
		schemaRef, err := f.fromTypeExpr(t, file, currPath, tag, ts.Elt, tps)
		if err != nil {
			return nil, newError(s.Pos(), err)
		}
		if title != "" || desc != "" || len(enums) > 0 && schemaRef.Ref != "" {
			s := openapi3.NewSchema().WithEnum(enums...)
			s.AllOf = openapi3.SchemaRefs{schemaRef}
			s.Title = title
			s.Description = desc
			return array(NewRef("", s), true), nil
		}
		return array(schemaRef, true), nil
	case *ast.StructType: // type x = struct{...}
		schema := openapi3.NewObjectSchema().WithEnum(enums...)
		schema.Title = title
		schema.Description = desc

		tps := buildTypeParams(s.TypeParams, tpRefs)
		if err := f.addFields(t, file, schema, currPath, tag, ts.Fields.List, tps); err != nil {
			return nil, err
		}

		return NewRef("", schema), nil
	case *ast.IndexExpr: // type x = G[int]
		return f.fromIndexExprType(t, file, currPath, tag, ts)
	case *ast.IndexListExpr: // type x = G[int, float]
		return f.fromIndexListExprType(t, file, currPath, tag, ts)
	default:
		msg := web.Phrase("%s can not convert to ast.StructType", s.Name.Name)
		return nil, newError(s.Pos(), msg)
	}
}

// 将 fields 中的所有字段解析到 schema
//
// 字段名如果存在 json 时，取 json 名称，否则直接采用字段名，xml 仅采用了 attr 和 parent>child 两种格式。
func (f SearchFunc) addFields(t *openapi.OpenAPI, file *ast.File, s *openapi3.Schema, modPath, tagName string, fields []*ast.Field, tpRefs map[string]*typeParam) error {
LOOP:
	for _, field := range fields {
		if len(field.Names) == 0 { // 嵌套对象
			ref, err := f.fromTypeExpr(t, file, modPath, tagName, field.Type, nil)
			if err != nil {
				return err
			}

			for k, v := range ref.Value.Properties {
				s.WithPropertyRef(k, v)
			}
			continue
		}

		// XMLName 特殊处理
		if field.Names[0].Name == "XMLName" && field.Tag != nil {
			structTag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
			tag := structTag.Get("xml")
			if tag == "-" || tag == "" { // 忽略此字段
				continue
			}

			items := strings.SplitN(tag, ",", 2)
			if items[0] != "" {
				s.XML = &openapi3.XML{Name: items[0]}
			}
			continue
		}

		name, nullable, xml := parseTag(field, tagName)
		if name == "-" {
			continue LOOP
		}

		item, err := f.fromTypeExpr(t, file, modPath, tagName, field.Type, tpRefs)
		if err != nil {
			return err
		}

		title, desc := parseComment(field.Comment, field.Doc)
		s.WithPropertyRef(name, wrap(item, title, desc, xml, nullable))
	}

	return nil
}

// 将 ast.Expr 中的内容转换到 schema 上
func (f SearchFunc) fromTypeExpr(t *openapi.OpenAPI, file *ast.File, currPath, tag string, e ast.Expr, tpRefs map[string]*typeParam) (*Ref, error) {
	switch expr := e.(type) {
	case *ast.ArrayType:
		schema, err := f.fromTypeExpr(t, file, currPath, tag, expr.Elt, tpRefs)
		if err != nil {
			return nil, err
		}
		return array(schema, true), nil
	case *ast.MapType: // NOTE: map 无法指定字段名
		return NewRef("", openapi3.NewObjectSchema()), nil
	case *ast.Ident: // int 或是 T
		if len(tpRefs) > 0 { // 这是泛型类型
			if elem, found := tpRefs[expr.Name]; found {
				return elem.ref, nil
			}
		}
		return f.fromName(t, currPath, expr.Name, tag, false)
	case *ast.StarExpr: // 指针 *Type
		return f.fromTypeExpr(t, file, currPath, tag, expr.X, tpRefs)
	case *ast.SelectorExpr: // json.Decoder
		mod, name := getSelectorExprName(expr, file)
		ref, err := f.fromName(t, currPath, mod+"."+name, tag, false) // , tpRefs
		if err != nil {
			var serr *Error
			if errors.As(err, &serr) {
				return nil, serr
			}
			return nil, newError(e.Pos(), err)
		}
		return ref, nil
	case *ast.StructType: // struct{...}
		s := openapi3.NewObjectSchema()
		if err := f.addFields(t, file, s, currPath, tag, expr.Fields.List, tpRefs); err != nil {
			return nil, err
		}
		return NewRef("", s), nil
	case *ast.IndexExpr: // Type[int] 或是 Type[T]
		return f.fromIndexExpr(t, file, currPath, tag, expr, tpRefs)
	case *ast.IndexListExpr: // Type[T, int]
		return f.fromIndexListExpr(t, file, currPath, tag, expr, tpRefs)
	//case *ast.InterfaceType: // 无法处理此类型
	default:
		return nil, newError(e.Pos(), web.Phrase("unsupported ast expr %+v", expr))
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
					if index := strings.IndexByte(words[0], '>'); index > 0 {
						xml = &openapi3.XML{Wrapped: true, Name: words[0][index+1:]}
					}
				case 2:
					wrapIndex := strings.IndexByte(words[0], '>')
					attr := words[1] == "attr"
					if wrapIndex > 0 || attr {
						xml = &openapi3.XML{Wrapped: wrapIndex > 0, Attribute: attr}
						if xml.Wrapped {
							xml.Name = words[0][wrapIndex+1:]
						}
					}
				}
			}
		}
	}
	return
}
