// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"golang.org/x/text/message"

	"github.com/issue9/web"
)

type Parameter struct {
	Ref *Ref

	Name       string
	Deprecated bool
	Required   bool
	Desc       web.LocaleStringer // 当前参数的描述信息

	// 以下字段是针对关联的 [Schema] 对象的。
	Type  string
	Title web.LocaleStringer
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
}

func (t *Parameter) buildParameter(p *message.Printer, in string) *renderer[parameterRenderer] {
	if t.Ref != nil {
		return newRenderer[parameterRenderer](t.Ref.build(p, "parameters"), nil)
	}

	return newRenderer(nil, &parameterRenderer{
		Name:        t.Name,
		In:          in,
		Required:    t.Required,
		Deprecated:  t.Deprecated,
		Description: sprint(p, t.Desc),
		Schema: newRenderer(nil, &schemaRenderer{
			Type:  t.Type,
			Title: sprint(p, t.Title),
		}),
	})
}

func (t *Parameter) buildHeader(p *message.Printer) *renderer[headerRenderer] {
	if t.Ref != nil {
		return newRenderer[headerRenderer](t.Ref.build(p, "headers"), nil)
	}

	return newRenderer(nil, &headerRenderer{
		Required:   t.Required,
		Deprecated: t.Deprecated,
	})
}

type Schema struct {
	Ref *Ref

	XML          *XML
	ExternalDocs *ExternalDocs
	Title        web.LocaleStringer
	Type         string
	AllOf        []*Schema
	OneOf        []*Schema
	AnyOf        []*Schema
	Format       string
	Items        *Schema
	Properties   []*Schema
}

type schemaRenderer struct {
	Type         string                      `json:"type" yaml:"type"`
	XML          *XML                        `json:"xml,omitempty" yaml:"xml,omitempty"`
	ExternalDocs *externalDocsRenderer       `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Title        string                      `json:"title,omitempty" yaml:"title,omitempty"`
	Description  string                      `json:"description,omitempty" yaml:"description,omitempty"`
	AllOf        []*renderer[schemaRenderer] `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	OneOf        []*renderer[schemaRenderer] `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	AnyOf        []*renderer[schemaRenderer] `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	Format       string                      `json:"format,omitempty" yaml:"format,omitempty"`
	Items        *renderer[schemaRenderer]   `json:"items,omitempty" yaml:"items,omitempty"`
	Properties   []*renderer[schemaRenderer] `json:"properties,omitempty" yaml:"properties,omitempty"`
}

func (s *Schema) addComponents(c *components) {
	if s.Ref == nil {
		return
	}

	if _, found := c.schemas[s.Ref.Ref]; !found {
		c.schemas[s.Ref.Ref] = s
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

	if s.Type == "" {
		panic("Type 不能为空")
	}
	return newRenderer(nil, &schemaRenderer{
		XML:          s.XML.clone(),
		ExternalDocs: s.ExternalDocs.build(p),
		Title:        sprint(p, s.Title),
		Type:         s.Type,
		AllOf:        cloneSchemas2SchemasRenderer(s.AllOf, p),
		OneOf:        cloneSchemas2SchemasRenderer(s.OneOf, p),
		AnyOf:        cloneSchemas2SchemasRenderer(s.AnyOf, p),
		Format:       s.Format,
		Items:        s.Items.build(p),
		Properties:   cloneSchemas2SchemasRenderer(s.Properties, p),
	})
}

func cloneSchemas2SchemasRenderer(s []*Schema, p *message.Printer) []*renderer[schemaRenderer] {
	ss := make([]*renderer[schemaRenderer], 0, len(s))
	for _, sss := range s {
		ss = append(ss, sss.build(p))
	}
	return ss
}
