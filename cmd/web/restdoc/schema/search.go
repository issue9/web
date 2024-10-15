// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package schema

import (
	"context"
	"fmt"
	"go/types"
	"reflect"
	"strconv"
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
// 如果 typePath 以 #components/schemas 开头，则从 [openapi3.T.Components.Schemas] 下查找。
// q 是否用于查询参数，如果是查询参数，那么字段名的获取将采用 json 且不会保存整个对象至 [openapi3.T.Components.Schemas] 之下。
//
// 可能返回的错误值为 [Error]
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

	if typ == nil { // typ == nil 也是正确的值，表示空值。
		return nil, nil
	}

	ref, _, err := s.fromType(t, "", typ, "", tag)
	return ref, err
}

// 从类型 typ 中构建 [openapi3.SchemaRef] 类型
//
// structRef 如果 typ 是 [pkg.Struct] 类型，此值用于指定该对象最终在 openapi 中的 ref 值。
func (s *Schema) fromType(t *openapi.OpenAPI, xmlName string, typ types.Type, structRef string, tag string) (ref *openapi3.SchemaRef, basic bool, err error) {
	if ref, ok := buildBasicType(typ); ok {
		return ref, true, nil
	}

	switch tt := typ.(type) {
	case *types.Pointer:
		ref, _, err = s.fromType(t, "", tt.Elem(), structRef, tag)
		if err != nil {
			return nil, false, err
		}
		return openapi.NewNullableSchemaRef(ref), false, nil
	case *types.Array:
		ref, _, err = s.fromType(t, "", tt.Elem(), structRef, tag)
		if err != nil {
			return nil, false, err
		}
		return openapi.NewArraySchemaRef(ref), false, nil
	case *types.Slice:
		ref, _, err = s.fromType(t, "", tt.Elem(), structRef, tag)
		if err != nil {
			return nil, false, err
		}
		return openapi.NewArraySchemaRef(ref), false, nil
	case *types.Basic:
		schemaRef, ok := buildBasicType(tt)
		if !ok {
			return nil, false, web.NewLocaleError("%s is not a valid basic type", tt.Name())
		}
		return schemaRef, true, nil
	case *pkg.Struct:
		if sref := s.getStruct(structRef, tt); sref != nil {
			return sref, false, nil
		}

		schema := openapi3.NewObjectSchema()
		if err := s.fromStruct(schema, t, xmlName, tt, tag); err != nil {
			return nil, false, err
		}
		return openapi.NewSchemaRef("", schema), false, nil
	case *pkg.Alias:
		refID := refReplacer.Replace(tt.ID())
		if schemaRef, found := t.GetSchema(refID); found { // 查找是否已经存在于 components/schemes
			return schemaRef, false, nil // basic 不会存在于 components/schemas
		}

		title, desc := parseComment(tt.Doc())
		docTypeName, docEnums, err := parseTypeDoc(tt.Obj(), title, desc)
		if err != nil {
			return nil, false, err
		}
		if docTypeName != "" { // 用户通过 @type 自定义了类型
			schema, err := buildSchema(openapi3.NewSchema(), docTypeName, docEnums...)
			if err != nil {
				return nil, false, err
			}
			schema.Title = title
			schema.Description = desc
			ref = openapi.NewSchemaRef(refID, schema)
		} else { // 指向其它类型，比如 type x = y 或是 type x struct {...} 等
			if xmlName == "" {
				xmlName = tt.Obj().Name()
			}
			ref, basic, err = s.fromType(t, xmlName, tt.Rhs(), refID, tag)
			if err != nil {
				return nil, false, err
			}
			if !basic && ref.Value != nil { // 不是从 *pkg.Struct 获得的数据
				ref = openapi.NewDocSchemaRef(ref, title, desc)
				if ref.Ref == "" {
					ref.Ref = refID
				}
			}
		}

		if tag == query.Tag { // 查询参数不保存至 #components/schemas
			ref.Ref = "" // 也没有必要有 ref.Ref
		} else {
			t.AddSchema(ref)
		}
		return ref, ref.Ref == "", nil
	case pkg.NotFound:
		return nil, false, web.NewLocaleError("not found type %s", tt.String())
	default:
		panic(fmt.Sprintf("未处理的类型 %T:%+v", typ, typ))
	}
}

// xmlName 结构体名称，同时也会被当作 XML 根元素名称（会被 XMLName 字段改写）；
func (s *Schema) fromStruct(schema *openapi3.Schema, t *openapi.OpenAPI, xmlName string, st *pkg.Struct, tag string) error {
	embeds := make([]*types.Var, 0, 3)

	for i := range st.NumFields() {
		field := st.Field(i)

		if field.Embedded() {
			embeds = append(embeds, field)
			continue
		}

		if !field.Exported() { // 嵌入对象名小写是可以的，所以要在 filed.Embedded 判断之后。
			continue
		}

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

		fieldRef, _, err := s.fromType(t, "", field.Type(), "", tag)
		if err != nil {
			return err
		}

		title, desc := parseComment(st.FieldDoc(i))
		name, nullable, xml := parseTag(field.Name(), tagValue, tag)
		schema.WithPropertyRef(name, openapi.NewAttrSchemaRef(fieldRef, title, desc, xml, nullable))
	}

	// 嵌入对象在最后执行，防止重名字段的冲突。
	for _, field := range embeds {
		fieldRef, _, err := s.fromType(t, "", field.Type(), "", tag)
		if err != nil {
			return err
		}

		if fieldRef.Value != nil && fieldRef.Value.Type.Is(openapi3.TypeObject) {
			for k, v := range fieldRef.Value.Properties {
				if _, found := schema.Properties[k]; found { // 防止与现有的重名
					continue
				}
				schema.WithPropertyRef(k, v)
			}
		}
	}

	return nil
}

func buildSchema(s *openapi3.Schema, docTypeName string, docEnums ...string) (*openapi3.Schema, error) {
	if docTypeName != "" {
		s.Type = &openapi3.Types{docTypeName}
	}

	if len(docEnums) == 0 {
		return s, nil
	}

	var enums []any
	if docTypeName == openapi3.TypeNumber || docTypeName == openapi3.TypeInteger {
		for _, val := range docEnums {
			v, err := strconv.Atoi(val)
			if err != nil {
				return nil, err
			}
			enums = append(enums, v)
		}
	} else {
		for _, val := range docEnums {
			enums = append(enums, val)
		}
	}

	s.WithEnum(enums...)
	return s, nil
}
