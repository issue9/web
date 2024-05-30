// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

const ComponentSchemaPrefix = "#/components/schemas/"

// RefEqual 判断两个 SchemaRef.Ref 是否相等
func RefEqual(r1, r2 string) bool {
	return strings.TrimPrefix(r1, ComponentSchemaPrefix) == strings.TrimPrefix(r2, ComponentSchemaPrefix)
}

// AddSchema 尝试添加一个 Schema 至 Components 中
//
// NOTE:  仅在 schema.Ref 不为空 或是 schema.Value 不为空时才会保存，且会对不规则的 schema.Ref 进行修正。
func (doc *OpenAPI) AddSchema(schema *openapi3.SchemaRef) {
	ref := strings.TrimPrefix(schema.Ref, ComponentSchemaPrefix)
	if ref == "" || schema.Value == nil {
		return
	}
	schema.Ref = ComponentSchemaPrefix + ref // 同时也统一 schema.Ref

	doc.schemaLocker.Lock()
	defer doc.schemaLocker.Unlock()
	if s, found := doc.doc.Components.Schemas[ref]; found {
		if reflect.DeepEqual(schema.Value, s.Value) {
			return
		}
		panic(fmt.Sprintf("添加同名的对象 %s，但是值不相同:\n%+v\n%+v", ref, schema.Value, s.Value))
	}

	doc.doc.Components.Schemas[ref] = NewSchemaRef("", schema.Value) // 防止保存时只写入 Ref
}

// GetSchema 从 Components 中查找 ref 引用的 Schema 定义
func (doc *OpenAPI) GetSchema(ref string) (*openapi3.SchemaRef, bool) {
	id := strings.TrimPrefix(ref, ComponentSchemaPrefix)

	doc.schemaLocker.RLock()
	defer doc.schemaLocker.RUnlock()
	if rr, found := doc.doc.Components.Schemas[id]; found {
		return NewSchemaRef(id, rr.Value), true
	}
	return nil, false
}

// NewSchemaRef 声明 [openapi3.SchemaRef]
//
// 在 ref 不为空且没有 [ComponentSchemaPrefix] 作为前缀时会自动添加前缀，其它情况下则不改变 ref 的值。
func NewSchemaRef(ref string, s *openapi3.Schema) *openapi3.SchemaRef {
	if ref != "" && !strings.HasPrefix(ref, ComponentSchemaPrefix) {
		ref = ComponentSchemaPrefix + ref
	}
	return openapi3.NewSchemaRef(ref, s)
}

// NewArraySchemaRef 将 ref 包装成数组
func NewArraySchemaRef(ref *openapi3.SchemaRef) *openapi3.SchemaRef {
	s := openapi3.NewArraySchema()
	s.Items = ref
	return NewSchemaRef("", s)
}

// NewDocSchemaRef 将 ref 附带上文档信息
//
// 这会以 AllOf 的形式形成一个新 [openapi3.SchemaRef] 对象，原有的 ref 不会被破坏。
func NewDocSchemaRef(ref *openapi3.SchemaRef, title, desc string) *openapi3.SchemaRef {
	if ref.Ref == "" { // 非引用模式，表示该值仅调用方使用，直接修改值。
		ref.Value.Title = title
		ref.Value.Description = desc
		return ref
	}

	if title != "" || desc != "" {
		s := openapi3.NewSchema()
		s.AllOf = openapi3.SchemaRefs{ref}
		if desc != "" {
			s.Description = desc
		}
		if title != "" {
			s.Title = title
		}
		return NewSchemaRef("", s)
	}

	return ref
}

// NewNullableSchemaRef 将 ref 包装为一个允许为空的对象
func NewNullableSchemaRef(ref *openapi3.SchemaRef) *openapi3.SchemaRef {
	if ref.Ref == "" { // 非引用模式，表示该值仅调用方使用，直接修改值。
		ref.Value.Nullable = true
		return ref
	}

	s := openapi3.NewSchema().WithNullable()
	s.AllOf = openapi3.SchemaRefs{ref}
	return NewSchemaRef("", s)
}

func NewAttrSchemaRef(ref *openapi3.SchemaRef, title, desc string, xml *openapi3.XML, nullable bool) *openapi3.SchemaRef {
	if ref == nil || (title == "" && desc == "" && xml == nil && !nullable) {
		return ref
	}

	var s *openapi3.Schema
	if ref.Ref == "" {
		s = ref.Value
	} else {
		s = openapi3.NewSchema()
		s.AllOf = openapi3.SchemaRefs{ref}
	}

	if title != "" {
		s.Title = title
	}
	if desc != "" {
		s.Description = desc
	}
	if xml != nil {
		s.XML = xml
	}
	if nullable {
		s.WithNullable()
	}

	return NewSchemaRef("", s)
}
