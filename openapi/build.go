// SPDX-FileCopyrightText: 2024-2025 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"maps"
	"net/http"
	"slices"
	"strings"
	"time"

	orderedmap "github.com/wk8/go-ordered-map/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/web"
)

// 此对象不能和 operationRenderer 放在同一文件，且需要在其之前编译，
// 否则会和造成泛型对象相互引用，无法编译。
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

func (d *Document) build(p *message.Printer, lang language.Tag, filterTags []string) *openAPIRenderer {
	servers := make([]*serverRenderer, 0, len(d.servers))
	for _, s := range d.servers {
		servers = append(servers, s.build(p))
	}

	security := make([]*orderedmap.OrderedMap[string, []string], 0, len(d.security))
	for _, sec := range d.security {
		pair := orderedmap.Pair[string, []string]{Key: sec.Name, Value: sec.Scopes}
		security = append(security, orderedmap.New[string, []string](orderedmap.WithInitialData(pair)))
	}

	tags := make([]*tagRenderer, 0, len(d.tags))
	for _, t := range d.tags {
		// NOTE: 标签不过滤，接口可能引用多个标签。
		tags = append(tags, t.build(p))
	}

	return &openAPIRenderer{
		OpenAPI:      Version,
		Info:         d.info.build(p),
		Servers:      servers,
		Paths:        writeMap2OrderedMap(d.paths, nil, func(in *PathItem) *renderer[pathItemRenderer] { return in.build(p, d, filterTags) }),
		WebHooks:     writeMap2OrderedMap(d.webHooks, nil, func(in *PathItem) *renderer[pathItemRenderer] { return in.build(p, d, filterTags) }),
		Components:   d.components.build(p, d),
		Security:     security,
		Tags:         tags,
		ExternalDocs: d.externalDocs.build(p),

		XLogo:        d.logo,
		XAssets:      d.assetsURL,
		XLanguage:    lang.String(),
		XModified:    d.last.Format(time.RFC3339),
		templateName: d.templateName,
	}
}

func (m *components) build(p *message.Printer, d *Document) *componentsRenderer {
	l := len(m.paths) + len(m.cookies) + len(m.queries)
	parameters := orderedmap.New[string, *parameterRenderer](orderedmap.WithCapacity[string, *parameterRenderer](l))
	writeMap2OrderedMap(m.paths, parameters, func(in *Parameter) *parameterRenderer { return in.buildParameterRenderer(p, InPath) })
	writeMap2OrderedMap(m.cookies, parameters, func(in *Parameter) *parameterRenderer { return in.buildParameterRenderer(p, InCookie) })
	writeMap2OrderedMap(m.queries, parameters, func(in *Parameter) *parameterRenderer { return in.buildParameterRenderer(p, InQuery) })

	return &componentsRenderer{
		Schemas:         writeMap2OrderedMap(m.schemas, nil, func(in *Schema) *schemaRenderer { return in.buildRenderer(p) }),
		Responses:       writeMap2OrderedMap(m.responses, nil, func(in *Response) *responseRenderer { return in.buildRenderer(p, d) }),
		Requests:        writeMap2OrderedMap(m.requests, nil, func(in *Request) *requestRenderer { return in.buildRenderer(p, d) }),
		SecuritySchemes: writeMap2OrderedMap(m.securitySchemes, nil, func(in *SecurityScheme) *securitySchemeRenderer { return in.build(p) }),
		Callbacks:       writeMap2OrderedMap(m.callbacks, nil, func(in *Callback) *callbackRenderer { return in.buildRenderer(p, d) }),
		PathItems:       writeMap2OrderedMap(m.pathItems, nil, func(in *PathItem) *pathItemRenderer { return in.buildRenderer(p, d, nil) }),

		Headers:    writeMap2OrderedMap(m.headers, nil, func(in *Parameter) *headerRenderer { return in.buildHeaderRenderer() }),
		Parameters: parameters,
	}
}

func (i *info) build(p *message.Printer) *infoRenderer {
	if i == nil {
		return nil
	}

	return &infoRenderer{
		Title:          sprint(p, i.title),
		Summary:        sprint(p, i.summary),
		Description:    sprint(p, i.description),
		TermsOfService: i.termsOfService,
		Contact:        i.contact.clone(),
		License:        i.license.clone(),
		Version:        i.version,
	}
}

func (t *tag) build(p *message.Printer) *tagRenderer {
	if t == nil {
		return nil
	}

	return &tagRenderer{
		Name:         t.name,
		Description:  sprint(p, t.description),
		ExternalDocs: t.externalDocs.build(p),
	}
}

func (e *ExternalDocs) build(p *message.Printer) *externalDocsRenderer {
	if e == nil {
		return nil
	}

	return &externalDocsRenderer{
		Description: sprint(p, e.Description),
		URL:         e.URL,
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
		Callbacks:    writeMap2OrderedMap(o.Callbacks, nil, func(in *Callback) *renderer[callbackRenderer] { return in.build(p, d) }),
		Security:     security,
		Servers:      servers,
		ExternalDocs: o.ExternalDocs.build(p),
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
		keys := slices.Collect(maps.Keys(resp.Content))
		slices.Sort(keys)
		for _, k := range keys {
			val := resp.Content[k] // 下面会修改 k，所以先获取值。

			if resp.Problem {
				if mt, found := d.mediaTypes[k]; found { // 找到 problem 专有的 mimetype
					k = mt
				}
			}
			content.Set(k, &mediaTypeRenderer{Schema: val.build(p)})
		}
	}
	if resp.Body != nil {
		for k, v := range d.mediaTypes {
			if resp.Problem {
				k = v
			}
			if content.GetPair(k) == nil {
				content.Set(k, &mediaTypeRenderer{
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
		for mt := range d.mediaTypes {
			if content.GetPair(mt) == nil {
				content.Set(mt, &mediaTypeRenderer{
					Schema: req.Body.build(p),
				})
			}
		}
	}

	return &requestRenderer{
		Content:     content,
		Required:    !req.Ignorable,
		Description: sprint(p, req.Description),
	}
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
		case http.MethodGet:
			path.Get = o.build(p, d)
		case http.MethodPut:
			path.Put = o.build(p, d)
		case http.MethodPost:
			path.Post = o.build(p, d)
		case http.MethodDelete:
			path.Delete = o.build(p, d)
		case http.MethodOptions:
			path.Options = o.build(p, d)
		case http.MethodHead:
			path.Head = o.build(p, d)
		case http.MethodPatch:
			path.Patch = o.build(p, d)
		case http.MethodTrace:
			path.Trace = o.build(p, d)
		}
	}

	return path
}

// 生成可用于渲染的对象
//
// typ 表示在 components 中的名称
func (ref *Ref) build(p *message.Printer, typ string) *refRenderer {
	if ref.Ref == "" {
		panic("ref 不能为空")
	}

	return &refRenderer{
		Ref:         "#/components/" + typ + "/" + ref.Ref,
		Summary:     sprint(p, ref.Summary),
		Description: sprint(p, ref.Description),
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
		Default:              s.Default,
	}
}

func cloneSchemas2SchemasRenderer(s []*Schema, p *message.Printer) []*renderer[schemaRenderer] {
	ss := make([]*renderer[schemaRenderer], 0, len(s))
	for _, sss := range s {
		ss = append(ss, sss.build(p))
	}
	return ss
}

func (e *SecurityScheme) build(p *message.Printer) *securitySchemeRenderer {
	return &securitySchemeRenderer{
		// NOTE: SecurityScheme.ID 作为 components.securitySchemes 的键名使用，并不出现在 securitySchemeRenderer
		Type:             e.Type,
		Description:      sprint(p, e.Description),
		Name:             e.Name,
		In:               e.In,
		Scheme:           e.Scheme,
		BearerFormat:     e.BearerFormat,
		Flows:            e.Flows.build(p),
		OpenIDConnectURL: e.OpenIDConnectURL,
	}
}

func (f *OAuthFlows) build(p *message.Printer) *oauthFlowsRenderer {
	if f == nil {
		return nil
	}

	return &oauthFlowsRenderer{
		Implicit:          f.Implicit.build(p),
		Password:          f.Password.build(p),
		ClientCredentials: f.ClientCredentials.build(p),
		AuthorizationCode: f.AuthorizationCode.build(p),
	}
}

func (f *OAuthFlow) build(p *message.Printer) *oauthFlowRenderer {
	if f == nil {
		return nil
	}

	return &oauthFlowRenderer{
		AuthorizationURL: f.AuthorizationURL,
		TokenUrl:         f.TokenURL,
		RefreshUrl:       f.RefreshURL,
		Scopes:           writeMap2OrderedMap(f.Scopes, nil, func(in web.LocaleStringer) string { return sprint(p, in) }),
	}
}

func (c *Callback) build(p *message.Printer, d *Document) *renderer[callbackRenderer] {
	if c.Ref != nil {
		return newRenderer[callbackRenderer](c.Ref.build(p, "callbacks"), nil)
	}
	return newRenderer(nil, c.buildRenderer(p, d))
}

func (c *Callback) buildRenderer(p *message.Printer, d *Document) *callbackRenderer {
	return writeMap2OrderedMap(c.Callback, nil, func(in *PathItem) *renderer[pathItemRenderer] { return in.build(p, d, nil) })
}
