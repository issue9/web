// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package openapi 采用 [web.Middleware] 中间件的形式生成 [openapi] 文档
//
// [openapi]: https://spec.openapis.org/oas/v3.1.1.html
package openapi

import (
	"crypto/md5"
	"encoding/hex"
	"slices"
	"strconv"
	"strings"
	"time"

	orderedmap "github.com/wk8/go-ordered-map/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/html"
	"github.com/issue9/web/mimetype/json"
	"github.com/issue9/web/mimetype/yaml"
)

// Document openapi 文档
type Document struct {
	info         *info
	servers      []*Server
	paths        map[string]*PathItem
	webHooks     map[string]*PathItem
	components   *components
	security     []*SecurityRequirement
	tags         []*tag
	externalDocs *ExternalDocs

	// 以下是一些预定义的项，不存在于 openAPIRenderer。

	mediaTypes    map[string]string // 所有接口都支持的类型，mimetype=>problem mimetype
	responses     map[string]string // key 为状态码，比如 4XX，值为 components 中的键名
	headers       []string          // components 中的键名
	cookies       []string          // components 中的键名
	enableOptions bool
	enableHead    bool

	// 与 HTML 模板相关的定义

	templateName string
	assetsURL    string
	favicon      string

	// 其它一些状态的设置

	disable bool      // 是否禁用
	last    time.Time // 最后向当前对象添加内容的时间，用于计算 ETag 值。

	s web.Server
}

type openAPIRenderer struct {
	OpenAPI      string                                                      `json:"openapi" yaml:"openapi"`
	Info         *infoRenderer                                               `json:"info" yaml:"info"`
	Servers      []*serverRenderer                                           `json:"servers,omitempty" yaml:"servers,omitempty"`
	Paths        *orderedmap.OrderedMap[string, *renderer[pathItemRenderer]] `json:"paths,omitempty" yaml:"paths,omitempty"`
	WebHooks     *orderedmap.OrderedMap[string, *renderer[pathItemRenderer]] `json:"webhooks,omitempty" yaml:"webhooks,omitempty"`
	Components   *componentsRenderer                                         `json:"components,omitempty" yaml:"components,omitempty"`
	Security     []*orderedmap.OrderedMap[string, []string]                  `json:"security,omitempty" yaml:"security,omitempty"`
	Tags         []*tagRenderer                                              `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs *externalDocsRenderer                                       `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`

	// 扩展内容

	XFavicon     string `json:"x-favicon,omitempty" yaml:"x-favicon,omitempty"`
	XAssets      string `json:"x-assets,omitempty" yaml:"x-assets,omitempty"`
	XLanguage    string `json:"x-language,omitempty" yaml:"x-language,omitempty"`
	XModified    string `json:"x-modified,omitempty" yaml:"x-modified,omitempty"`
	templateName string
}

// New 声明 [Document] 对象
//
// title 文档的标题；
func New(s web.Server, title web.LocaleStringer, o ...Option) *Document {
	doc := &Document{
		info: &info{
			title:   title,
			version: s.Version(),
		},
		components: newComponents(),

		mediaTypes: make(map[string]string, 5),
		responses:  make(map[string]string, 5),

		last: time.Now(),

		s: s,
	}

	for _, opt := range o {
		opt(doc)
	}

	return doc
}

type documentQuery struct {
	Tags []string `query:"tag"`
}

// Handler 实现 [web.HandlerFunc] 接口
//
// 目前支持以下几种格式：
//   - json 通过将 accept 报头设置为 [json.Mimetype] 返回 JSON 格式的数据；
//   - yaml 通过将 accept 报头设置为 [yaml.Mimetype] 返回 YAML 格式的数据；
//   - html 通过将 accept 报头设置为 [html.Mimetype] 返回 HTML 格式的数据。
//     需要通过 [WithHTML] 进行配置，可参考 [github.com/issue9/web/mimetype/html]；
//
// NOTE: Handler 支持的输出格式限定在以上几种，但是最终是否能正常输出以上几种格式，
// 还需要由 [web.Server] 是否配置相应的解码方式。
//
// 该路由接受 tag 查询参数，在未指定参数的情况下，表示返回所有接口，
// 如果指定了参数，则只返回带指定标签的接口，多个标签以逗号分隔。
func (d *Document) Handler(ctx *web.Context) web.Responser {
	if d.disable {
		return ctx.NotImplemented()
	}

	q := &documentQuery{}
	if resp := ctx.QueryObject(true, q, web.ProblemBadRequest); resp != nil {
		return resp
	}

	if m := ctx.Mimetype(false); (m != json.Mimetype && m != yaml.Mimetype && m != html.Mimetype) ||
		(m == html.Mimetype && d.templateName == "") {
		return ctx.Problem(web.ProblemNotAcceptable)
	}

	dataURL := ctx.Request().URL.Path
	if len(q.Tags) > 0 {
		dataURL += "?tag=" + strings.Join(q.Tags, ",")
	}

	return web.NotModified(func() (string, bool) {
		slices.Sort(q.Tags)

		// 引起 ETag 变化的几个要素
		etag := strconv.Itoa(int(d.last.Unix())) + "/" +
			strings.Join(q.Tags, ",") + "/" +
			ctx.Mimetype(false) + "/" +
			ctx.LanguageTag().String()
		h := md5.New()
		h.Write([]byte(etag))
		val := h.Sum(nil)
		return hex.EncodeToString(val), true
	}, func() (any, error) {
		return d.build(ctx.LocalePrinter(), ctx.LanguageTag(), q.Tags), nil
	})
}

func (o *openAPIRenderer) MarshalHTML() (name string, data any) {
	return o.templateName, o
}

// Disable 是否禁用 [Document.Handler] 接口输出内容
func (d *Document) Disable(disable bool) { d.disable = disable }

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
		if len(filterTags) > 0 && slices.Index(filterTags, t.name) >= 0 {
			tags = append(tags, t.build(p))
		} else {
			tags = append(tags, t.build(p))
		}
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

		XFavicon:     d.favicon,
		XAssets:      d.assetsURL,
		XLanguage:    lang.String(),
		XModified:    d.last.Format(time.RFC3339),
		templateName: d.templateName,
	}
}

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

type info struct {
	title          web.LocaleStringer
	summary        web.LocaleStringer
	description    web.LocaleStringer
	termsOfService string
	contact        *contactRender
	license        *licenseRenderer
	version        string
}

type infoRenderer struct {
	Title          string           `json:"title" yaml:"title"`
	Summary        string           `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description    string           `json:"description,omitempty" yaml:"description,omitempty"`
	TermsOfService string           `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Contact        *contactRender   `json:"contact,omitempty" yaml:"contact,omitempty"`
	License        *licenseRenderer `json:"license,omitempty" yaml:"license,omitempty"`
	Version        string           `json:"version" yaml:"version"`
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

type tag struct {
	name         string
	description  web.LocaleStringer
	externalDocs *ExternalDocs
}

type tagRenderer struct {
	Name         string                `json:"name" yaml:"name"`
	Description  string                `json:"description,omitempty" yaml:"description,omitempty"`
	ExternalDocs *externalDocsRenderer `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
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

type ExternalDocs struct {
	Description web.LocaleStringer
	URL         string
}

type externalDocsRenderer struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	URL         string `json:"url" yaml:"url"`
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

type XML struct {
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Prefix    string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Wrapped   bool   `json:"wrapped,omitempty" yaml:"wrapped,omitempty"`
	Attribute bool   `json:"attribute,omitempty" yaml:"attribute,omitempty"`
}

func (xml *XML) clone() *XML {
	if xml == nil {
		return nil
	}

	return &XML{
		Name:      xml.Name,
		Namespace: xml.Namespace,
		Prefix:    xml.Prefix,
		Wrapped:   xml.Wrapped,
		Attribute: xml.Attribute,
	}
}

type licenseRenderer struct {
	Name       string `json:"name" yaml:"name"`
	Identifier string `json:"identifier,omitempty" yaml:"identifier,omitempty"`
	URL        string `json:"url,omitempty" yaml:"url,omitempty"`
}

func newLicense(name string, id string) *licenseRenderer {
	switch {
	case id == "":
		return &licenseRenderer{Name: name}
	case strings.HasPrefix(id, "http://") || strings.HasPrefix(id, "https://"):
		return &licenseRenderer{Name: name, URL: id}
	default:
		return &licenseRenderer{Name: name, Identifier: id}
	}
}

func (l *licenseRenderer) clone() *licenseRenderer {
	if l == nil {
		return nil
	}

	return &licenseRenderer{
		Name:       l.Name,
		Identifier: l.Identifier,
		URL:        l.URL,
	}
}

type contactRender struct {
	Name  string `json:"name" yaml:"name"`
	URL   string `json:"url" yaml:"url"`
	Email string `json:"email" yaml:"email"`
}

func (c *contactRender) clone() *contactRender {
	if c == nil {
		return nil
	}

	return &contactRender{
		Name:  c.Name,
		Email: c.Email,
		URL:   c.URL,
	}
}
