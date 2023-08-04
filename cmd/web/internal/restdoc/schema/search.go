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
	"github.com/issue9/query/v3"
	"github.com/issue9/sliceutil"
	"github.com/issue9/web"

	"github.com/issue9/web/cmd/web/internal/restdoc/pkg"
)

type SearchFunc func(string) *pkg.Package

// currPath 当前包的导出路径；
// typePath 表示需要查找的类型名，非内置类型且不带路径信息，则将 currPath 作为路径信息；
// typePath 可以包含类型参数，比如 G[int]。
// q 是否用于查询参数
//
// 可能返回的错误值为 *Error
func (f SearchFunc) New(t *OpenAPI, currPath, typePath string, q bool) (*Ref, error) {
	var isArray bool
	if strings.HasPrefix(typePath, "[]") {
		typePath = typePath[2:]
		isArray = true
	}

	tag := "json"
	if q {
		tag = query.Tag
	}

	var tpRefs []*Ref
	if index := strings.LastIndexByte(typePath, '['); index > 0 && typePath[len(typePath)-1] == ']' {
		tps := strings.Split(typePath[index+1:len(typePath)-1], ",")

		tpRefs = make([]*Ref, 0, len(tps))
		for _, i := range tps {
			idxRef, err := f.fromName(t, currPath, strings.TrimSpace(i), tag, false, nil)
			if err != nil {
				return nil, err
			}
			tpRefs = append(tpRefs, idxRef)
		}
		//typePath = typePath[:index]
	}

	return f.fromName(t, currPath, typePath, tag, isArray, tpRefs)
}

// 根据类型名生成 schema 对象
//
// tpRefs 泛型参数对应的 *Ref，非泛型则为空；
// 其它参数参考 [SearchFunc.New]
func (f SearchFunc) fromName(t *OpenAPI, currPath, typePath, tag string, isArray bool, tpRefs []*Ref) (*Ref, error) {
	switch typePath { // 基本类型
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

	structName := typePath
	if index := strings.LastIndexByte(typePath, '.'); index > 0 { // 全局的路径
		currPath = typePath[:index]
		structName = typePath[index+1:]
	} else {
		typePath = currPath + "." + typePath
	}
	if currPath == "" {
		return nil, web.NewLocaleError("not found %s", typePath) // 行数未变化，直接返回错误。
	}

	ref := refReplacer.Replace(typePath)

	if index := strings.LastIndexByte(structName, '['); index > 0 {
		structName = structName[:index]
	}

	if schemaRef, found := t.Components.Schemas[ref]; found { // 查找是否已经存在于 components/schemes
		sr := NewRef(ref, schemaRef.Value)
		addRefPrefix(sr)
		return array(sr, isArray), nil
	}

	pkg := f(currPath)
	if pkg == nil {
		return nil, web.NewLocaleError("not found %s", currPath) // 行数未变化，直接返回错误。
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
				if spec, ok = s.(*ast.TypeSpec); ok && spec.Name.Name == structName {
					file = f
					break LOOP // 找到了，就退到最外层。
				}
			}
		}
	}

	if spec == nil || file == nil {
		return nil, web.NewLocaleError("not found %s", typePath)
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
func (f SearchFunc) fromTypeSpec(t *OpenAPI, file *ast.File, currPath, ref, tag string, s *ast.TypeSpec, tpRefs []*Ref) (*Ref, error) {
	title, desc, enums := parseTypeDoc(s)
	if desc == "" && s.Comment != nil {
		desc = s.Comment.Text()
	}

	switch ts := s.Type.(type) {
	case *ast.Ident: // type x = int 或是 type x int
		schemaRef, err := f.fromName(t, currPath, ts.Name, tag, false, nil)
		if err != nil {
			return nil, newError(s.Pos(), err)
		}
		schemaRef.Value.Title = title
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
		schemaRef.Value.Title = title
		schemaRef.Value.Description = desc
		schemaRef.Value.Enum = enums
		return schemaRef, nil
	case *ast.StructType: // type x = struct{...}
		schema := openapi3.NewObjectSchema()
		schema.Title = title
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

func (f SearchFunc) fromIndexExpr(t *OpenAPI, file *ast.File, currPath, tag string, idx *ast.IndexExpr) (*Ref, error) {
	mod, idxName := getExprName(file, currPath, idx.Index)
	idxRef, err := f.fromName(t, mod, idxName, tag, false, nil)
	if err != nil {
		return nil, err
	}

	mod, name := getExprName(file, currPath, idx.X)
	return f.fromName(t, mod, name, tag, false, []*Ref{idxRef})
}

func (f SearchFunc) fromIndexListExpr(t *OpenAPI, file *ast.File, currPath, ref, tag string, idx *ast.IndexListExpr) (*Ref, error) {
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
func (f SearchFunc) addFields(t *OpenAPI, file *ast.File, s *openapi3.Schema, modPath, tagName string, fields []*ast.Field, tp *ast.FieldList, tpRefs []*Ref) error {
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

		item, err := f.fromExpr(t, file, modPath, tagName, field.Type, tp, tpRefs)
		if err != nil {
			return err
		}

		title, desc := parseComment(field.Comment, field.Doc)
		s.WithPropertyRef(name, wrap(item, title, desc, xml, nullable))
	}

	return nil
}

// 将 ast.Expr 中的内容转换到 schema 上
func (f SearchFunc) fromExpr(t *OpenAPI, file *ast.File, currPath, tag string, e ast.Expr, tp *ast.FieldList, tpRefs []*Ref) (*Ref, error) {
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
	case *ast.StructType:
		s := openapi3.NewObjectSchema()
		if err := f.addFields(t, file, s, currPath, tag, expr.Fields.List, tp, tpRefs); err != nil {
			return nil, err
		}
		return NewRef("", s), nil
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
