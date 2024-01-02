// SPDX-License-Identifier: MIT

package schema

import (
	"context"
	"fmt"
	"go/types"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/query/v3"
	"github.com/issue9/web"

	"github.com/issue9/web/cmd/web/restdoc/openapi"
	"github.com/issue9/web/cmd/web/restdoc/pkg"
)

// New 根据类型名称 typePath 生成 SchemaRef 对象
//
// typePath 表示需要查找的类型名，一般格式为 [path.]name，path 为包的路径，name 为类型名。
// 如果是内置类型可省略 path 部分。
// 如果 typePath 以 #components/schemas 开头，则从 t.Components.Schemas 下查找。
// q 是否用于查询参数，如果是查询参数，那么字段名的获取将采用 json。
//
// 可能返回的错误值为 *Error
func (s *Schema) New(ctx context.Context, t *openapi.OpenAPI, typePath string, q bool) (*openapi3.SchemaRef, error) {
	tag := "json"
	if q { // 查询参数采用 query.Tag 获取为字段名
		tag = query.Tag
	}

	if strings.HasPrefix(typePath, openapi.ComponentSchemaPrefix) {
		if ref, found := t.GetSchema(typePath); found {
			return ref, nil // basic 不会存在于 components/schemas
		} else {
			return nil, web.NewLocaleError("not found openapi schema ref %s", typePath)
		}
	}

	typ, err := s.Packages().TypeOf(ctx, typePath)
	if err != nil {
		return nil, err
	}

	ref, _, err := s.fromType(t, "", typ, tag)
	return ref, err
}

// 从类型 typ 中构建 [Ref] 类型
//
// xmlName typ 为 [pkg.Struct] 时，为其指定的最外层 xml 名称；
// tps 为与 typ 对应的范型参数列表；
func (s *Schema) fromType(t *openapi.OpenAPI, xmlName string, typ types.Type, tag string) (ref *openapi3.SchemaRef, basic bool, err error) {
	if ref, ok := buildBasicType(typ.String()); ok {
		return ref, true, nil
	}

	switch tt := typ.(type) {
	case *types.Pointer: // 指针
		ref, basic, err = s.fromType(t, "", tt.Elem(), tag)
		if err != nil {
			return nil, false, err
		}
		return openapi.NewNullableSchemaRef(ref), basic, nil
	case *types.Array:
		ref, basic, err = s.fromType(t, "", tt.Elem(), tag)
		if err != nil {
			return nil, false, err
		}
		return openapi.NewArraySchemaRef(ref), basic, nil
	case *types.Slice:
		ref, basic, err = s.fromType(t, "", tt.Elem(), tag)
		if err != nil {
			return nil, false, err
		}
		return openapi.NewArraySchemaRef(ref), basic, nil
	case *types.Basic:
		schemaRef, ok := buildBasicType(tt.Name())
		if !ok {
			return nil, false, web.NewLocaleError("%s is not a valid basic type", tt.Name())
		}
		return schemaRef, true, nil
	case *pkg.Struct:
		ref, err := s.fromStruct(t, xmlName, tt, tag)
		return ref, false, err
	case *pkg.Named:
		refID := refReplacer.Replace(tt.ID())
		if schemaRef, found := t.GetSchema(refID); found { // 查找是否已经存在于 components/schemes
			return schemaRef, false, nil // basic 不会存在于 components/schemas
		}

		title, desc := parseComment(tt.Doc())
		docTypeName, docEnums := parseTypeDoc(title, desc)
		if docTypeName != "" { // 用户通过 @type 自定义了类型
			schema := buildSchema(openapi3.NewSchema(), docTypeName, docEnums...)
			schema.Title = title
			schema.Description = desc
			ref = openapi3.NewSchemaRef(refID, schema)
		} else { // 指向其它类型，比如 type x = y 或是 type x struct {...} 等，指向了 y 的定义
			if xmlName == "" {
				xmlName = tt.Obj().Name()
			}
			ref, basic, err = s.fromType(t, xmlName, tt.Next(), tag)
			if err != nil {
				return nil, false, err
			}
			ref = openapi.NewDocSchemaRef(ref, title, desc)
			if !basic {
				ref.Ref = refID
			}
		}

		if tag != query.Tag { // 查询参数不保存整个对象
			t.AddSchema(ref)
		}
		return ref, ref.Ref == "", nil
	default:
		panic(fmt.Sprintf("未处理的类型 %T:%+v", typ, typ))
	}
}

// xmlName 表示结构体作为 xml 对象时的根元素名称，如果结构体中包含了 XMLName 字段，会改写 xmlName 的值；
// 将 *pkg.Struct 解析为 schema 对象
func (s *Schema) fromStruct(t *openapi.OpenAPI, xmlName string, st *pkg.Struct, tag string) (*openapi3.SchemaRef, error) {
	schema := openapi3.NewObjectSchema()

	// BUG 结构体与嵌入字段重名的处理

	for i := 0; i < st.NumFields(); i++ {
		field := st.Field(i)
		// 匿名和非导出字段由 pkg.TypeOf 过滤

		tagValue := st.Tag(i)

		if field.Name() == "XMLName" { // XML 的特殊处理
			xmlValue := reflect.StructTag(strings.Trim(tagValue, "`")).Get("xml")
			switch xmlValue {
			case "-": // 忽略
			case "":
				schema.XML = &openapi3.XML{Name: xmlName}
			default:
				if items := strings.SplitN(xmlValue, ",", 2); items[0] != "" {
					schema.XML = &openapi3.XML{Name: items[0]}
				} else {
					schema.XML = &openapi3.XML{Name: xmlName}
				}
			}
			continue
		}

		fieldRef, _, err := s.fromType(t, "", field.Type(), tag)
		if err != nil {
			return nil, err
		}

		title, desc := parseComment(st.FieldDoc(i))
		name, nullable, xml := parseTag(field.Name(), tagValue, tag)
		schema.WithPropertyRef(name, openapi.NewAttrSchemaRef(fieldRef, title, desc, xml, nullable))
	}

	return openapi.NewSchemaRef("", schema), nil
}

func buildSchema(s *openapi3.Schema, docTypeName string, docEnums ...any) *openapi3.Schema {
	if len(docEnums) > 0 {
		s.WithEnum(docEnums...)
	}

	if docTypeName != "" {
		s.Type = docTypeName
	}

	return s
}
