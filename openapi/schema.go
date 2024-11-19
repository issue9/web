// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"

	orderedmap "github.com/wk8/go-ordered-map/v2"
	"golang.org/x/text/message"

	"github.com/issue9/web"
)

type Parameter struct {
	Ref *Ref

	Name        string
	Deprecated  bool
	Required    bool
	Description web.LocaleStringer // 当前参数的描述信息

	// InHeader 模式下此值无效
	Schema *Schema
}

type parameterRenderer struct {
	Name        string                    `json:"name" yaml:"name"`
	In          string                    `json:"in" yaml:"in"`
	Required    bool                      `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated  bool                      `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Schema      *renderer[schemaRenderer] `json:"schema,omitempty" yaml:"schema,omitempty"`
	Description string                    `json:"description,omitempty" yaml:"description,omitempty"`
}

type headerRenderer struct {
	Required   bool `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated bool `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
}

// skipRefNotNil 当存在 ref 时忽略内容的检测
func (t *Parameter) valid(skipRefNotNil bool) *web.FieldError {
	if skipRefNotNil && t.Ref != nil {
		return nil
	}

	if t.Name == "" {
		return web.NewFieldError("Name", "不能为空")
	}

	if t.Schema == nil {
		return web.NewFieldError("Schema", "不能为空")
	}

	if err := t.Schema.valid(skipRefNotNil); err != nil {
		err.AddFieldParent("Schema")
		return err
	}

	if t.Schema.Type == TypeObject {
		return web.NewFieldError("Schema", "不支持复杂类型")
	}

	return nil
}

func (t *Parameter) addComponents(c *components, in string) {
	if t.Ref == nil {
		return
	}

	switch in {
	case InCookie:
		if _, found := c.cookies[t.Ref.Ref]; !found {
			c.cookies[t.Ref.Ref] = t
		}
	case InHeader:
		if _, found := c.headers[t.Ref.Ref]; !found {
			c.headers[t.Ref.Ref] = t
		}
	case InQuery:
		if _, found := c.queries[t.Ref.Ref]; !found {
			c.queries[t.Ref.Ref] = t
		}
	case InPath:
		if _, found := c.paths[t.Ref.Ref]; !found {
			c.paths[t.Ref.Ref] = t
		}
	}

	if t.Schema != nil {
		t.Schema.addComponents(c)
	}
}

func (t *Parameter) buildParameter(p *message.Printer, in string) *renderer[parameterRenderer] {
	if t.Ref != nil {
		return newRenderer[parameterRenderer](t.Ref.build(p, "parameters"), nil)
	}
	return newRenderer(nil, t.buildParameterRenderer(p, in))
}

func (t *Parameter) buildParameterRenderer(p *message.Printer, in string) *parameterRenderer {
	return &parameterRenderer{
		Name:        t.Name,
		In:          in,
		Required:    t.Required,
		Deprecated:  t.Deprecated,
		Description: sprint(p, t.Description),
		Schema:      t.Schema.build(p),
	}
}

func (t *Parameter) buildHeader(p *message.Printer) *renderer[headerRenderer] {
	if t.Ref != nil {
		return newRenderer[headerRenderer](t.Ref.build(p, "headers"), nil)
	}
	return newRenderer(nil, t.buildHeaderRenderer())
}

func (t *Parameter) buildHeaderRenderer() *headerRenderer {
	return &headerRenderer{Required: t.Required, Deprecated: t.Deprecated}
}

type Schema struct {
	Ref *Ref

	XML                  *XML
	ExternalDocs         *ExternalDocs
	Title                web.LocaleStringer
	Description          web.LocaleStringer
	Type                 string
	AllOf                []*Schema
	OneOf                []*Schema
	AnyOf                []*Schema
	Format               string
	Items                *Schema
	Properties           map[string]*Schema
	AdditionalProperties *Schema
	Required             []string
	Minimum              int
	Maximum              int
	Enum                 []any
}

type schemaRenderer struct {
	Type                 string                                                    `json:"type" yaml:"type"`
	XML                  *XML                                                      `json:"xml,omitempty" yaml:"xml,omitempty"`
	ExternalDocs         *externalDocsRenderer                                     `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Title                string                                                    `json:"title,omitempty" yaml:"title,omitempty"`
	Description          string                                                    `json:"description,omitempty" yaml:"description,omitempty"`
	AllOf                []*renderer[schemaRenderer]                               `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	OneOf                []*renderer[schemaRenderer]                               `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	AnyOf                []*renderer[schemaRenderer]                               `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	Format               string                                                    `json:"format,omitempty" yaml:"format,omitempty"`
	Items                *renderer[schemaRenderer]                                 `json:"items,omitempty" yaml:"items,omitempty"`
	Properties           *orderedmap.OrderedMap[string, *renderer[schemaRenderer]] `json:"properties,omitempty" yaml:"properties,omitempty"`
	AdditionalProperties *renderer[schemaRenderer]                                 `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Required             []string                                                  `json:"required,omitempty" yaml:"required,omitempty"`
	Minimum              int                                                       `json:"minimum,omitempty" yaml:"minimum,omitempty"`
	Maximum              int                                                       `json:"maximum,omitempty" yaml:"maximum,omitempty"`
	Enum                 []any                                                     `json:"enum,omitempty" yaml:"enum,omitempty"`
}

func (d *Document) newSchema(t reflect.Type) *Schema {
	return schemaFromType(d, t, true, "", nil)
}

// NewSchema 根据 [reflect.Type] 生成 [Schema] 对象
func NewSchema(t reflect.Type, title, desc web.LocaleStringer) *Schema {
	s := schemaFromType(nil, t, true, "", desc)
	s.Title = title
	return s
}

// d 仅用于查找其关联的 components/schema 中是否存在相同名称的对象，如果存在则直接生成引用对象。
//
// desc 表示类型 t 的 Description 属性
// rootName 根结构体的名称，主要是为了解决子元素又引用了根元素的类型引起的循环引用。
func schemaFromType(d *Document, t reflect.Type, isRoot bool, rootName string, desc web.LocaleStringer) *Schema {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return &Schema{Type: TypeString, Description: desc}
	case reflect.Bool:
		return &Schema{Type: TypeBoolean, Description: desc}
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: TypeNumber, Description: desc}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &Schema{Type: TypeInteger, Description: desc}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Schema{Type: TypeInteger, Minimum: 0, Description: desc}
	case reflect.Array, reflect.Slice:
		return &Schema{Type: TypeArray, Items: schemaFromType(d, t.Elem(), false, rootName, nil), Description: desc}
	case reflect.Map:
		return &Schema{Type: TypeObject, AdditionalProperties: schemaFromType(d, t.Elem(), false, rootName, nil), Description: desc}
	case reflect.Struct:
		return schemaFromObject(d, t, isRoot, rootName, desc)
	}
	return nil
}

func schemaFromObject(d *Document, t reflect.Type, isRoot bool, rootName string, desc web.LocaleStringer) *Schema {
	typeName := getTypeName(t)

	if d != nil {
		if s, found := d.components.schemas[typeName]; found { // 已经存在于 components
			return s
		}
	}

	ref := &Ref{Ref: typeName}
	if isRoot {
		rootName = typeName // isRoot == true 时，rootName 必然为空
	} else if typeName == rootName { // 在字段中引用了根对象
		return &Schema{Ref: ref}
	}

	ps := make(map[string]*Schema, t.NumField())
	req := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		k := f.Type.Kind()
		var itemTitle web.LocaleStringer

		if f.IsExported() && k != reflect.Chan && k != reflect.Func && k != reflect.Complex64 && k != reflect.Complex128 {
			name := f.Name
			var xml *XML
			if f.Tag != "" {
				tag, omitempty := getTagName(f, "json")
				if tag == "-" {
					continue
				} else if tag != "" {
					name = tag
				}

				if !omitempty {
					req = append(req, name)
				}

				if xmlName, _ := getTagName(f, "xml"); xmlName != "" && xmlName != name {
					xml = &XML{Name: xmlName}
				}

				comment := f.Tag.Get(CommentTag)
				if comment != "" {
					itemTitle = web.Phrase(comment)
				}
			}

			s := schemaFromType(d, t.Field(i).Type, false, rootName, itemTitle)
			if s == nil {
				continue
			}

			if xml != nil {
				s.XML = xml
			}
			ps[name] = s
		}
	}

	return &Schema{
		Type:        TypeObject,
		Properties:  ps,
		Ref:         ref,
		Required:    req,
		Description: desc,
	}
}

// skipRefNotNil 当存在 ref 时忽略内容的检测
func (s *Schema) valid(skipRefNotNil bool) *web.FieldError {
	if skipRefNotNil && s.Ref != nil {
		return nil
	}

	if s.Type == "" {
		return web.NewFieldError("Type", "不能为空")
	}

	for i, item := range s.AllOf {
		if err := item.valid(skipRefNotNil); err != nil {
			err.AddFieldParent("AllOf[" + strconv.Itoa(i) + "]")
			return err
		}
	}

	for i, item := range s.AnyOf {
		if err := item.valid(skipRefNotNil); err != nil {
			err.AddFieldParent("AnyOf[" + strconv.Itoa(i) + "]")
			return err
		}
	}

	for i, item := range s.OneOf {
		if err := item.valid(skipRefNotNil); err != nil {
			err.AddFieldParent("OneOf[" + strconv.Itoa(i) + "]")
			return err
		}
	}

	if s.Items != nil {
		if err := s.Items.valid(skipRefNotNil); err != nil {
			err.AddFieldParent("Items")
			return err
		}
		if s.Type != TypeArray {
			return web.NewFieldError("Type", fmt.Sprintf("必须为 %s", TypeArray))
		}
	}
	if s.Type == TypeArray && s.Items == nil {
		return web.NewFieldError("Type", fmt.Sprintf("必须为 %s", TypeArray))
	}

	for key, item := range s.Properties {
		if err := item.valid(skipRefNotNil); err != nil {
			err.AddFieldParent("Properties[" + key + "]")
			return err
		}
	}
	if len(s.Properties) > 0 && s.Type != TypeObject {
		return web.NewFieldError("Type", fmt.Sprintf("必须为 %s", TypeObject))
	}

	if s.AdditionalProperties != nil {
		if err := s.AdditionalProperties.valid(skipRefNotNil); err != nil {
			err.AddFieldParent("AdditionalProperties")
			return err
		}
	}

	return nil
}

func (s *Schema) addComponents(c *components) {
	if s.Ref != nil {
		if _, found := c.schemas[s.Ref.Ref]; !found {
			c.schemas[s.Ref.Ref] = s
		}
	}

	for _, item := range s.AllOf {
		item.addComponents(c)
	}

	for _, item := range s.OneOf {
		item.addComponents(c)
	}

	for _, item := range s.AnyOf {
		item.addComponents(c)
	}

	if s.Items != nil {
		s.Items.addComponents(c)
	}

	for _, item := range s.Properties {
		item.addComponents(c)
	}
}

func (s *Schema) build(p *message.Printer) *renderer[schemaRenderer] {
	if s == nil {
		return nil
	}

	if s.Ref != nil {
		return newRenderer[schemaRenderer](s.Ref.build(p, "schemas"), nil)
	}
	return newRenderer(nil, s.buildRenderer(p))
}

func (s *Schema) buildRenderer(p *message.Printer) *schemaRenderer {
	return &schemaRenderer{
		XML:                  s.XML.clone(),
		ExternalDocs:         s.ExternalDocs.build(p),
		Title:                sprint(p, s.Title),
		Description:          sprint(p, s.Description),
		Type:                 s.Type,
		AllOf:                cloneSchemas2SchemasRenderer(s.AllOf, p),
		OneOf:                cloneSchemas2SchemasRenderer(s.OneOf, p),
		AnyOf:                cloneSchemas2SchemasRenderer(s.AnyOf, p),
		Format:               s.Format,
		Items:                s.Items.build(p),
		Properties:           writeMap2OrderedMap(s.Properties, nil, func(in *Schema) *renderer[schemaRenderer] { return in.build(p) }),
		AdditionalProperties: s.AdditionalProperties.build(p),
		Required:             s.Required,
		Minimum:              s.Minimum,
		Maximum:              s.Maximum,
		Enum:                 slices.Clone(s.Enum),
	}
}

func cloneSchemas2SchemasRenderer(s []*Schema, p *message.Printer) []*renderer[schemaRenderer] {
	ss := make([]*renderer[schemaRenderer], 0, len(s))
	for _, sss := range s {
		ss = append(ss, sss.build(p))
	}
	return ss
}
