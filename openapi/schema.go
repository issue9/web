// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"reflect"

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
}

func (d *Document) newSchema(t reflect.Type) *Schema {
	return schemaFromType(d, t, true, "")
}

// NewSchema 根据 [reflect.Type] 生成 [Schema] 对象
func NewSchema(t reflect.Type) *Schema {
	return schemaFromType(nil, t, true, "")
}

// d 仅用于查找其关联的 components/schema 中是否存在相同名称的对象，如果存在则直接生成引用对象。
func schemaFromType(d *Document, t reflect.Type, isRoot bool, rootName string) *Schema {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return &Schema{Type: TypeString}
	case reflect.Bool:
		return &Schema{Type: TypeBoolean}
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: TypeNumber}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &Schema{Type: TypeInteger}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Schema{Type: TypeInteger, Minimum: 0}
	case reflect.Array, reflect.Slice:
		return &Schema{Type: TypeArray, Items: schemaFromType(d, t.Elem(), false, rootName)}
	case reflect.Map:
		return &Schema{Type: TypeObject, AdditionalProperties: schemaFromType(d, t.Elem(), false, rootName)}
	case reflect.Struct:
		return schemaFromObject(d, t, isRoot, rootName)
	}
	return nil
}

func schemaFromObject(d *Document, t reflect.Type, isRoot bool, rootName string) *Schema {
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
			}

			s := schemaFromType(d, t.Field(i).Type, false, rootName)
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
		Type:       TypeObject,
		Properties: ps,
		Ref:        ref,
		Required:   req,
	}
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
	}
}

func cloneSchemas2SchemasRenderer(s []*Schema, p *message.Printer) []*renderer[schemaRenderer] {
	ss := make([]*renderer[schemaRenderer], 0, len(s))
	for _, sss := range s {
		ss = append(ss, sss.build(p))
	}
	return ss
}
