// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import orderedmap "github.com/wk8/go-ordered-map/v2"

type components struct {
	schemas         map[string]*Schema
	responses       map[string]*Response
	requests        map[string]*Request
	securitySchemes map[string]*SecurityScheme
	callbacks       map[string]*Callback
	pathItems       map[string]*PathItem

	paths   map[string]*Parameter // 路径中的参数
	queries map[string]*Parameter // 查询参数
	headers map[string]*Parameter // 该值写入 components/headers 之下，而不是 components/parameters
	cookies map[string]*Parameter
}

type componentsRenderer struct {
	Schemas         *orderedmap.OrderedMap[string, *schemaRenderer]         `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	Responses       *orderedmap.OrderedMap[string, *responseRenderer]       `json:"responses,omitempty" yaml:"responses,omitempty"`
	Parameters      *orderedmap.OrderedMap[string, *parameterRenderer]      `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Requests        *orderedmap.OrderedMap[string, *requestRenderer]        `json:"requestBodies,omitempty" yaml:"requestBodies,omitempty"`
	Headers         *orderedmap.OrderedMap[string, *headerRenderer]         `json:"headers,omitempty" yaml:"headers,omitempty"`
	SecuritySchemes *orderedmap.OrderedMap[string, *securitySchemeRenderer] `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
	Callbacks       *orderedmap.OrderedMap[string, *callbackRenderer]       `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
	PathItems       *orderedmap.OrderedMap[string, *pathItemRenderer]       `json:"pathItems,omitempty" yaml:"pathItems,omitempty"`
}

func newComponents() *components {
	return &components{
		schemas:         make(map[string]*Schema, 100),
		responses:       make(map[string]*Response),
		requests:        make(map[string]*Request),
		securitySchemes: make(map[string]*SecurityScheme),
		callbacks:       make(map[string]*Callback),
		pathItems:       make(map[string]*PathItem),

		paths:   make(map[string]*Parameter),
		queries: make(map[string]*Parameter),
		headers: make(map[string]*Parameter),
		cookies: make(map[string]*Parameter),
	}
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

func (resp *Callback) addComponents(c *components) {
	if resp.Ref != nil {
		if _, found := c.callbacks[resp.Ref.Ref]; !found {
			c.callbacks[resp.Ref.Ref] = resp
		}
	}

	for _, item := range resp.Callback {
		item.addComponents(c)
	}
}

// 所有带 $ref 的字段如果还未存在于 c，则会写入。
func (o *Operation) addComponents(c *components) {
	for _, p := range o.Paths {
		p.addComponents(c, InPath)
	}

	for _, p := range o.Queries {
		p.addComponents(c, InQuery)
	}

	for _, p := range o.Cookies {
		p.addComponents(c, InCookie)
	}

	for _, p := range o.Headers {
		p.addComponents(c, InHeader)
	}

	if o.RequestBody != nil {
		o.RequestBody.addComponents(c)
	}

	for _, r := range o.Responses {
		r.addComponents(c)
	}

	for _, r := range o.Callbacks {
		r.addComponents(c)
	}
}

func (resp *Response) addComponents(c *components) {
	if resp.Ref != nil {
		if _, found := c.responses[resp.Ref.Ref]; !found {
			c.responses[resp.Ref.Ref] = resp
		}
	}

	for _, h := range resp.Headers {
		h.addComponents(c, InHeader)
	}

	if resp.Body != nil {
		resp.Body.addComponents(c)
	}

	for _, s := range resp.Content {
		s.addComponents(c)
	}
}

func (req *Request) addComponents(c *components) {
	if req.Ref != nil {
		if _, found := c.requests[req.Ref.Ref]; !found {
			c.requests[req.Ref.Ref] = req
		}
	}

	if req.Body != nil {
		req.Body.addComponents(c)
	}

	if len(req.Content) > 0 {
		for _, s := range req.Content {
			s.addComponents(c)
		}
	}
}

func (item *PathItem) addComponents(c *components) {
	if item.Ref != nil {
		if _, found := c.pathItems[item.Ref.Ref]; !found {
			c.pathItems[item.Ref.Ref] = item
		}
	}

	for _, p := range item.Paths {
		p.addComponents(c, InPath)
	}
	for _, p := range item.Queries {
		p.addComponents(c, InQuery)
	}
	for _, p := range item.Headers {
		p.addComponents(c, InHeader)
	}
	for _, p := range item.Cookies {
		p.addComponents(c, InCookie)
	}
}
