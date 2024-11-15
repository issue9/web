// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"fmt"
	"slices"
	"strings"

	orderedmap "github.com/wk8/go-ordered-map/v2"
	"golang.org/x/text/message"

	"github.com/issue9/web"
)

// Operation 定义了每一个 API 的属性
type Operation struct {
	Summary     web.LocaleStringer
	Description web.LocaleStringer
	ID          string
	Deprecated  bool
	Tags        []string

	Paths        []*Parameter // 路径中的参数
	Queries      []*Parameter // 查询参数
	Headers      []*Parameter
	Cookies      []*Parameter
	RequestBody  *Request
	Responses    map[int]*Response // key = 状态码
	Callbacks    map[string]Callback
	Security     []*SecurityRequirement
	Servers      []*Server
	ExternalDocs *ExternalDocs
}

type operationRenderer struct {
	Tags         []string                                                    `json:"tags,omitempty" yaml:"tags,omitempty"`
	Summary      string                                                      `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                                                      `json:"description,omitempty" yaml:"description,omitempty"`
	ID           string                                                      `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Deprecated   bool                                                        `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Parameters   []*renderer[parameterRenderer]                              `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody  *renderer[requestRenderer]                                  `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses    *orderedmap.OrderedMap[int, *renderer[responseRenderer]]    `json:"responses,omitempty" yaml:"responses,omitempty"`
	Callbacks    *orderedmap.OrderedMap[string, *renderer[callbackRenderer]] `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
	Security     []*securityRequirementRenderer                              `json:"security,omitempty" yaml:"security,omitempty"`
	Servers      []*serverRenderer                                           `json:"servers,omitempty" yaml:"servers,omitempty"`
	ExternalDocs *externalDocsRenderer                                       `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
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

func (o *Operation) build(p *message.Printer, d *Document) *operationRenderer {
	if o == nil {
		return nil
	}

	parameters := make([]*renderer[parameterRenderer], 0, len(o.Paths)+len(o.Cookies)+len(o.Headers)+len(o.Queries))
	for _, param := range o.Paths {
		parameters = append(parameters, param.buildParameter(p, InPath))
	}
	for _, param := range o.Cookies {
		parameters = append(parameters, param.buildParameter(p, InCookie))
	}
	for _, param := range o.Headers {
		parameters = append(parameters, param.buildParameter(p, InHeader))
	}
	for _, param := range o.Queries {
		parameters = append(parameters, param.buildParameter(p, InQuery))
	}

	security := make([]*orderedmap.OrderedMap[string, []string], 0, len(o.Security))
	for _, sec := range o.Security {
		pair := orderedmap.Pair[string, []string]{Key: sec.Name, Value: sec.Scopes}
		security = append(security, orderedmap.New[string, []string](orderedmap.WithInitialData(pair)))
	}

	servers := make([]*serverRenderer, 0, len(o.Servers))
	for _, s := range o.Servers {
		servers = append(servers, s.build(p))
	}

	return &operationRenderer{
		Tags:         slices.Clone(o.Tags),
		Summary:      sprint(p, o.Summary),
		Description:  sprint(p, o.Description),
		ID:           o.ID,
		Deprecated:   o.Deprecated,
		Parameters:   parameters,
		RequestBody:  o.RequestBody.build(p, d),
		Responses:    writeMap2OrderedMap(o.Responses, nil, func(in *Response) *renderer[responseRenderer] { return in.build(p, d) }),
		Callbacks:    writeMap2OrderedMap(o.Callbacks, nil, func(in Callback) *renderer[callbackRenderer] { return in.build(p, d) }),
		Security:     security,
		Servers:      servers,
		ExternalDocs: o.ExternalDocs.build(p),
	}
}

type Callback struct {
	Ref      *Ref
	Callback map[string]*PathItem
}

type callbackRenderer = orderedmap.OrderedMap[string, *renderer[pathItemRenderer]]

func (c *Callback) build(p *message.Printer, d *Document) *renderer[callbackRenderer] {
	if c.Ref != nil {
		return newRenderer[callbackRenderer](c.Ref.build(p, "callbacks"), nil)
	}
	return newRenderer(nil, c.buildRenderer(p, d))
}

func (c *Callback) buildRenderer(p *message.Printer, d *Document) *callbackRenderer {
	return writeMap2OrderedMap(c.Callback, nil, func(in *PathItem) *renderer[pathItemRenderer] { return in.build(p, d, nil) })
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

type Response struct {
	Ref         *Ref
	Headers     []*Parameter
	Description web.LocaleStringer

	// Body 和 Content 共同组成了正文内容
	// 所有不在 Content 中出现的类型均采用 [openAPI.MediaTypesRenderer] 与 Body 相结构。
	Body    *Schema
	Content map[string]*Schema // key = mimetype
}

type responseRenderer struct {
	Description string                                                    `json:"description" yaml:"description"`
	Headers     *orderedmap.OrderedMap[string, *renderer[headerRenderer]] `json:"headers,omitempty" yaml:"headers,omitempty"`
	Content     *orderedmap.OrderedMap[string, *mediaTypeRenderer]        `json:"content,omitempty" yaml:"content,omitempty"`
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

func (resp *Response) build(p *message.Printer, d *Document) *renderer[responseRenderer] {
	if resp == nil {
		return nil
	}

	if resp.Ref != nil {
		return newRenderer[responseRenderer](resp.Ref.build(p, "responses"), nil)
	}
	return newRenderer(nil, resp.buildRenderer(p, d))
}

func (resp *Response) buildRenderer(p *message.Printer, d *Document) *responseRenderer {
	var headers *orderedmap.OrderedMap[string, *renderer[headerRenderer]]
	if resp.Headers != nil {
		headers = orderedmap.New[string, *renderer[headerRenderer]](orderedmap.WithCapacity[string, *headerRenderer](len(resp.Headers)))
		for _, h := range resp.Headers {
			headers.Set(h.Name, h.buildHeader(p))
		}
	}

	content := orderedmap.New[string, *mediaTypeRenderer](orderedmap.WithCapacity[string, *mediaTypeRenderer](len(d.mediaTypes)))
	if resp.Content != nil {
		writeMap2OrderedMap(resp.Content, content, func(in *Schema) *mediaTypeRenderer {
			return &mediaTypeRenderer{Schema: in.build(p)}
		})
	}
	if resp.Body != nil {
		for _, mt := range d.mediaTypes {
			if content.GetPair(mt) == nil {
				content.Set(mt, &mediaTypeRenderer{
					Schema: resp.Body.build(p),
				})
			}
		}
	}

	return &responseRenderer{
		Description: sprint(p, resp.Description),
		Headers:     headers,
		Content:     content,
	}
}

type Request struct {
	Ref       *Ref
	Ignorable bool // 对应 requestBody.required

	// Body 和 Content 共同组成了正文内容
	// 所有不在 Content 中出现的类型均采用 [Document.MediaTypes] 与 Body 相结构。
	Body    *Schema
	Content map[string]*Schema // key = mimetype
}

type requestRenderer struct {
	Content  *orderedmap.OrderedMap[string, *mediaTypeRenderer] `json:"content" yaml:"content"`
	Required bool                                               `json:"required,omitempty" yaml:"required,omitempty"`
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

	for _, s := range req.Content {
		s.addComponents(c)
	}
}

func (req *Request) build(p *message.Printer, d *Document) *renderer[requestRenderer] {
	if req == nil {
		return nil
	}

	if req.Ref != nil {
		return newRenderer[requestRenderer](req.Ref.build(p, "requestBodies"), nil)
	}
	return newRenderer(nil, req.buildRenderer(p, d))
}

func (req *Request) buildRenderer(p *message.Printer, d *Document) *requestRenderer {
	content := orderedmap.New[string, *mediaTypeRenderer](orderedmap.WithCapacity[string, *mediaTypeRenderer](len(d.mediaTypes)))
	if req.Content != nil {
		writeMap2OrderedMap(req.Content, content, func(in *Schema) *mediaTypeRenderer {
			return &mediaTypeRenderer{Schema: in.build(p)}
		})
	}
	if req.Body != nil {
		for _, mt := range d.mediaTypes {
			if content.GetPair(mt) == nil {
				content.Set(mt, &mediaTypeRenderer{
					Schema: req.Body.build(p),
				})
			}
		}
	}

	return &requestRenderer{
		Content:  content,
		Required: !req.Ignorable,
	}
}

type Server struct {
	URL         string
	Description web.LocaleStringer
	Variables   []*ServerVariable
}

type serverRenderer struct {
	URL         string                                                  `json:"url" yaml:"url"`
	Description string                                                  `json:"description,omitempty" yaml:"description,omitempty"`
	Variables   *orderedmap.OrderedMap[string, *serverVariableRenderer] `json:"variables,omitempty" yaml:"variables,omitempty"`
}

type ServerVariable struct {
	Name        string
	Default     string
	Description web.LocaleStringer
	Enums       []string
}

type serverVariableRenderer struct {
	Enum        []string `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default     string   `json:"default" yaml:"default"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
}

func (s *Server) valid() *web.FieldError {
	params := getPathParams(s.URL)
	if len(params) != len(s.Variables) {
		return web.NewFieldError("Variables", "参数与路径中的指定不同")
	}

	for _, v := range s.Variables {
		if slices.Index(params, v.Name) < 0 {
			return web.NewFieldError("Variables", fmt.Sprintf("参数 %s 不存在于路径中", v.Name))
		}
	}

	return nil
}

func (s *Server) build(p *message.Printer) *serverRenderer {
	if s == nil {
		return nil
	}

	rs := &serverRenderer{
		URL:         s.URL,
		Description: sprint(p, s.Description),
	}

	if s.Variables != nil {
		rs.Variables = orderedmap.New[string, *serverVariableRenderer](orderedmap.WithCapacity[string, *serverVariableRenderer](len(s.Variables)))
		for _, v := range s.Variables {
			rs.Variables.Set(v.Name, &serverVariableRenderer{
				Enum:        v.Enums,
				Default:     v.Default,
				Description: sprint(p, v.Description),
			})
		}
	}

	return rs
}

type mediaTypeRenderer struct {
	Schema *renderer[schemaRenderer] `json:"schema,omitempty" yaml:"schema,omitempty"`
}

type PathItem struct {
	Ref        *Ref
	Operations map[string]*Operation
	Servers    []*Server

	Paths   []*Parameter // 路径中的参数
	Queries []*Parameter // 查询参数
	Headers []*Parameter
	Cookies []*Parameter
}

type pathItemRenderer struct {
	Get        *operationRenderer             `json:"get,omitempty" yaml:"get,omitempty"`
	Put        *operationRenderer             `json:"put,omitempty" yaml:"put,omitempty"`
	Post       *operationRenderer             `json:"post,omitempty" yaml:"post,omitempty"`
	Delete     *operationRenderer             `json:"delete,omitempty" yaml:"delete,omitempty"`
	Options    *operationRenderer             `json:"options,omitempty" yaml:"options,omitempty"`
	Head       *operationRenderer             `json:"head,omitempty" yaml:"head,omitempty"`
	Patch      *operationRenderer             `json:"patch,omitempty" yaml:"patch,omitempty"`
	Trace      *operationRenderer             `json:"trace,omitempty" yaml:"trace,omitempty"`
	Servers    []*serverRenderer              `json:"servers,omitempty" yaml:"servers,omitempty"`
	Parameters []*renderer[parameterRenderer] `json:"parameters,omitempty" yaml:"parameters,omitempty"`
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

func (item *PathItem) build(p *message.Printer, d *Document, tags []string) *renderer[pathItemRenderer] {
	if item == nil {
		return nil
	}

	if item.Ref != nil {
		return newRenderer[pathItemRenderer](item.Ref.build(p, "pathItems"), nil)
	}

	return newRenderer(nil, item.buildRenderer(p, d, tags))
}

func (item *PathItem) buildRenderer(p *message.Printer, d *Document, tags []string) *pathItemRenderer {
	parameters := make([]*renderer[parameterRenderer], 0, len(item.Paths)+len(item.Cookies)+len(item.Headers)+len(item.Queries))
	for _, param := range item.Paths {
		parameters = append(parameters, param.buildParameter(p, InPath))
	}
	for _, param := range item.Cookies {
		parameters = append(parameters, param.buildParameter(p, InCookie))
	}
	for _, param := range item.Headers {
		parameters = append(parameters, param.buildParameter(p, InHeader))
	}
	for _, param := range item.Queries {
		parameters = append(parameters, param.buildParameter(p, InQuery))
	}

	servers := make([]*serverRenderer, 0, len(item.Servers))
	for _, s := range item.Servers {
		servers = append(servers, s.build(p))
	}

	path := &pathItemRenderer{
		Servers:    servers,
		Parameters: parameters,
	}

	for method, o := range item.Operations {
		if len(tags) > 0 && slices.IndexFunc(o.Tags, func(t string) bool { return slices.Index(tags, t) >= 0 }) < 0 {
			continue
		}

		switch strings.ToUpper(method) {
		case "GET":
			path.Get = o.build(p, d)
		case "PUT":
			path.Put = o.build(p, d)
		case "POST":
			path.Post = o.build(p, d)
		case "DELETE":
			path.Delete = o.build(p, d)
		case "OPTIONS":
			path.Options = o.build(p, d)
		case "HEAD":
			path.Head = o.build(p, d)
		case "PATCH":
			path.Patch = o.build(p, d)
		case "TRACE":
			path.Trace = o.build(p, d)
		}
	}

	return path
}
