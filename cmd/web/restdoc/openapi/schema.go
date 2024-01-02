// SPDX-License-Identifier: MIT

package openapi

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

const ComponentSchemaPrefix = "#/components/schemas/"

// AddSchema 尝试添加一个 Schema 至 Components 中
//
// NOTE:  仅在 schema.Ref 不为空时才会保存。
func (doc *OpenAPI) AddSchema(schema *openapi3.SchemaRef) {
	// NOTE: Components.Schemas 的键名不包含 #/components/schemas/ 前缀，
	// 但是 SchemaRef.Ref 是包含此前缀的。

	ref := strings.TrimPrefix(schema.Ref, ComponentSchemaPrefix)
	if ref == "" {
		return
	}
	schema.Ref = ComponentSchemaPrefix + ref

	doc.schemaLocker.Lock()
	defer doc.schemaLocker.Unlock()
	if s, found := doc.doc.Components.Schemas[ref]; found {
		if schema.Value == s.Value {
			return
		}
		panic(fmt.Sprintf("添加同名的对象 %s", ref))
	}
	doc.doc.Components.Schemas[ref] = schema
}

// GetSchema 从 Components 中查找 ref 引用的 Schema 定义
func (doc *OpenAPI) GetSchema(ref string) (*openapi3.SchemaRef, bool) {
	ref = strings.TrimPrefix(ref, ComponentSchemaPrefix)

	doc.schemaLocker.RLock()
	defer doc.schemaLocker.RUnlock()
	r, found := doc.doc.Components.Schemas[ref]
	return r, found
}

func NewSchemaRef(refID string, s *openapi3.Schema) *openapi3.SchemaRef {
	if refID != "" && !strings.HasPrefix(refID, ComponentSchemaPrefix) {
		refID = ComponentSchemaPrefix + refID
	}
	return openapi3.NewSchemaRef(refID, s)
}

// NewArraySchemaRef 将 ref 包装成数组
func NewArraySchemaRef(ref *openapi3.SchemaRef) *openapi3.SchemaRef {
	s := openapi3.NewArraySchema()
	s.Items = ref
	return openapi3.NewSchemaRef("", s)
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

// NewXMLSchemaRef 包装 ref 为其指定 XML 属性
func NewXMLSchemaRef(ref *openapi3.SchemaRef, xml *openapi3.XML) *openapi3.SchemaRef {
	if ref.Ref == "" { // 非引用模式，表示该值仅调用方使用，直接修改值。
		ref.Value.XML = xml
		return ref
	}

	if xml != nil {
		s := openapi3.NewSchema()
		s.AllOf = openapi3.SchemaRefs{ref}
		s.XML = xml
		return NewSchemaRef("", s)
	}

	return ref
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
