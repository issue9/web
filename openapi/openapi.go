// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package openapi 采用 [web.Middleware] 中间件的形式生成 [openapi] 文档
//
// 结构体标签
//
// - comment 用于可翻译的注释，该内容会被翻译后保存在字段的 Schema.Description 中；
// - openapi 对 openapi 类型的自定义，格式为 type,format，可以自定义字段的类型和格式；
//
// [openapi]: https://spec.openapis.org/oas/v3.1.1.html
package openapi

import (
	"fmt"
	"time"

	orderedmap "github.com/wk8/go-ordered-map/v2"

	"github.com/issue9/web"
)

// Version 支持的 openapi 版本
const Version = "3.1.0"

// CommentTag 可提取翻译内容的结构体标签名称
const CommentTag = "comment"

const (
	InPath   = "path"
	InQuery  = "query"
	InHeader = "header"
	InCookie = "cookie"
)

const (
	TypeString  = "string"
	TypeNull    = "null"
	TypeBoolean = "boolean"
	TypeArray   = "array"
	TypeNumber  = "number"
	TypeInteger = "integer"
	TypeObject  = "object"
)

const (
	FormatInt32    = "int32"
	FormatInt64    = "int64"
	FormatFloat    = "float"
	FormatDouble   = "double"
	FormatPassword = "password"
	FormatDate     = "date"
	FormatTime     = "time"
	FormatDateTime = "date-time"
)

const (
	SecuritySchemeTypeHTTP          = "http"
	SecuritySchemeTypeAPIKey        = "apiKey"
	SecuritySchemeTypeMutualTLS     = "mutualTLS"
	SecuritySchemeTypeOAuth2        = "oauth2"
	SecuritySchemeTypeOpenIDConnect = "openIdConnect"
)

type (
	// Document openapi 文档
	Document struct {
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

	openAPIRenderer struct {
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

	// Ref 定义了 $ref
	Ref struct {
		Ref         string
		Summary     web.LocaleStringer
		Description web.LocaleStringer
	}

	refRenderer struct {
		Ref         string `json:"$ref" yaml:"$ref"`
		Summary     string `json:"summary,omitempty" yaml:"summary,omitempty"`
		Description string `json:"description,omitempty" yaml:"description,omitempty"`
	}

	info struct {
		title          web.LocaleStringer
		summary        web.LocaleStringer
		description    web.LocaleStringer
		termsOfService string
		contact        *contactRender
		license        *licenseRenderer
		version        string
	}

	infoRenderer struct {
		Title          string           `json:"title" yaml:"title"`
		Summary        string           `json:"summary,omitempty" yaml:"summary,omitempty"`
		Description    string           `json:"description,omitempty" yaml:"description,omitempty"`
		TermsOfService string           `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
		Contact        *contactRender   `json:"contact,omitempty" yaml:"contact,omitempty"`
		License        *licenseRenderer `json:"license,omitempty" yaml:"license,omitempty"`
		Version        string           `json:"version" yaml:"version"`
	}

	tag struct {
		name         string
		description  web.LocaleStringer
		externalDocs *ExternalDocs
	}

	tagRenderer struct {
		Name         string                `json:"name" yaml:"name"`
		Description  string                `json:"description,omitempty" yaml:"description,omitempty"`
		ExternalDocs *externalDocsRenderer `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	}

	ExternalDocs struct {
		Description web.LocaleStringer
		URL         string
	}

	externalDocsRenderer struct {
		Description string `json:"description,omitempty" yaml:"description,omitempty"`
		URL         string `json:"url" yaml:"url"`
	}

	XML struct {
		Name      string `json:"name,omitempty" yaml:"name,omitempty"`
		Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
		Prefix    string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
		Wrapped   bool   `json:"wrapped,omitempty" yaml:"wrapped,omitempty"`
		Attribute bool   `json:"attribute,omitempty" yaml:"attribute,omitempty"`
	}

	licenseRenderer struct {
		Name       string `json:"name" yaml:"name"`
		Identifier string `json:"identifier,omitempty" yaml:"identifier,omitempty"`
		URL        string `json:"url,omitempty" yaml:"url,omitempty"`
	}

	contactRender struct {
		Name  string `json:"name" yaml:"name"`
		URL   string `json:"url" yaml:"url"`
		Email string `json:"email" yaml:"email"`
	}

	Callback struct {
		Ref      *Ref
		Callback map[string]*PathItem
	}

	callbackRenderer = orderedmap.OrderedMap[string, *renderer[pathItemRenderer]]

	SecurityScheme struct {
		ID string // 在 components 中的键名，要求唯一性。

		Type             string
		Description      web.LocaleStringer
		Name             string
		In               string
		Scheme           string
		BearerFormat     string
		Flows            *OAuthFlows
		OpenIDConnectURL string
	}

	securitySchemeRenderer struct {
		Type             string              `json:"type" yaml:"type"`
		Description      string              `json:"description,omitempty" yaml:"description,omitempty"`
		Name             string              `json:"name,omitempty" yaml:"name,omitempty"`
		In               string              `json:"in,omitempty" yaml:"in,omitempty"`
		Scheme           string              `json:"scheme,omitempty" yaml:"scheme,omitempty"`
		BearerFormat     string              `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
		Flows            *oauthFlowsRenderer `json:"flows,omitempty" yaml:"flows,omitempty"`
		OpenIDConnectURL string              `json:"openIdConnectUrl,omitempty" yaml:"openIdConnectUrl,omitempty"`
	}

	OAuthFlows struct {
		Implicit          *OAuthFlow
		Password          *OAuthFlow
		ClientCredentials *OAuthFlow
		AuthorizationCode *OAuthFlow
	}

	oauthFlowsRenderer struct {
		Implicit          *oauthFlowRenderer `json:"implicit,omitempty" yaml:"implicit,omitempty"`
		Password          *oauthFlowRenderer `json:"password,omitempty" yaml:"password,omitempty"`
		ClientCredentials *oauthFlowRenderer `json:"clientCredentials,omitempty" yaml:"clientCredentials,omitempty"`
		AuthorizationCode *oauthFlowRenderer `json:"authorizationCode,omitempty" yaml:"authorizationCode,omitempty"`
	}

	OAuthFlow struct {
		AuthorizationURL string
		TokenURL         string
		RefreshURL       string
		Scopes           map[string]web.LocaleStringer
	}

	oauthFlowRenderer struct {
		AuthorizationURL string                                 `json:"authorizationUrl,omitempty" yaml:"authorizationUrl,omitempty"`
		TokenUrl         string                                 `json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
		RefreshUrl       string                                 `json:"refreshUrl,omitempty" yaml:"refreshUrl,omitempty"`
		Scopes           *orderedmap.OrderedMap[string, string] `json:"scopes,omitempty" yaml:"scopes,omitempty"`
	}

	SecurityRequirement struct {
		Name   string
		Scopes []string
	}

	securityRequirementRenderer = orderedmap.OrderedMap[string, []string]

	// Operation 定义了每一个 API 的属性
	Operation struct {
		d *Document

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
		Responses    map[string]*Response // key = 状态码，比如 2XX
		Callbacks    map[string]*Callback // key = 名称
		Security     []*SecurityRequirement
		Servers      []*Server
		ExternalDocs *ExternalDocs
	}

	operationRenderer struct {
		Tags        []string                                                    `json:"tags,omitempty" yaml:"tags,omitempty"`
		Summary     string                                                      `json:"summary,omitempty" yaml:"summary,omitempty"`
		Description string                                                      `json:"description,omitempty" yaml:"description,omitempty"`
		ID          string                                                      `json:"operationId,omitempty" yaml:"operationId,omitempty"`
		Deprecated  bool                                                        `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
		Parameters  []*renderer[parameterRenderer]                              `json:"parameters,omitempty" yaml:"parameters,omitempty"`
		RequestBody *renderer[requestRenderer]                                  `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
		Responses   *orderedmap.OrderedMap[string, *renderer[responseRenderer]] `json:"responses,omitempty" yaml:"responses,omitempty"`

		Callbacks    *orderedmap.OrderedMap[string, *renderer[callbackRenderer]] `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
		Security     []*securityRequirementRenderer                              `json:"security,omitempty" yaml:"security,omitempty"`
		Servers      []*serverRenderer                                           `json:"servers,omitempty" yaml:"servers,omitempty"`
		ExternalDocs *externalDocsRenderer                                       `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	}

	Response struct {
		Ref         *Ref
		Headers     []*Parameter
		Description web.LocaleStringer

		// 是否是用于表示错误类型的
		//
		// 该值如果为 true，会尝试在查找对应的媒体类型，
		// 比如在 [web.Server] 将 application/json 对应为 application/problem+json，
		// 那么在当前对象输出时，也会将媒体类型转换为 application/problem+json。
		Problem bool

		// Body 和 Content 共同组成了正文内容
		// 所有不在 Content 中出现的类型均采用 [openAPI.MediaTypesRenderer] 与 Body 相结合。
		Body    *Schema
		Content map[string]*Schema // key = mimetype
	}

	responseRenderer struct {
		Description string                                                    `json:"description" yaml:"description"`
		Headers     *orderedmap.OrderedMap[string, *renderer[headerRenderer]] `json:"headers,omitempty" yaml:"headers,omitempty"`
		Content     *orderedmap.OrderedMap[string, *mediaTypeRenderer]        `json:"content,omitempty" yaml:"content,omitempty"`
	}

	Request struct {
		Ref         *Ref
		Ignorable   bool // 对应 requestBody.required
		Description web.LocaleStringer

		// Body 和 Content 共同组成了正文内容
		// 所有不在 Content 中出现的类型均采用 [Document.MediaTypes] 与 Body 相结合。
		Body    *Schema
		Content map[string]*Schema // key = mimetype
	}

	requestRenderer struct {
		Content     *orderedmap.OrderedMap[string, *mediaTypeRenderer] `json:"content" yaml:"content"`
		Required    bool                                               `json:"required,omitempty" yaml:"required,omitempty"`
		Description string                                             `json:"description" yaml:"description"`
	}

	Server struct {
		URL         string
		Description web.LocaleStringer
		Variables   []*ServerVariable
	}

	serverRenderer struct {
		URL         string                                                  `json:"url" yaml:"url"`
		Description string                                                  `json:"description,omitempty" yaml:"description,omitempty"`
		Variables   *orderedmap.OrderedMap[string, *serverVariableRenderer] `json:"variables,omitempty" yaml:"variables,omitempty"`
	}

	ServerVariable struct {
		Name        string
		Default     string
		Description web.LocaleStringer
		Enums       []string
	}

	serverVariableRenderer struct {
		Enum        []string `json:"enum,omitempty" yaml:"enum,omitempty"`
		Default     string   `json:"default" yaml:"default"`
		Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	}

	mediaTypeRenderer struct {
		Schema *renderer[schemaRenderer] `json:"schema,omitempty" yaml:"schema,omitempty"`
	}

	PathItem struct {
		Ref        *Ref
		Operations map[string]*Operation
		Servers    []*Server

		Paths   []*Parameter // 路径中的参数
		Queries []*Parameter // 查询参数
		Headers []*Parameter
		Cookies []*Parameter
	}
)

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

// Disable 是否禁用 [Document.Handler] 接口输出内容
func (d *Document) Disable(disable bool) { d.disable = disable }

// AddWebhook 添加 Webhook 的定义
func (d *Document) AddWebhook(name, method string, o *Operation) {
	if d.webHooks == nil {
		d.webHooks = make(map[string]*PathItem, 5)
	}

	hook, found := d.webHooks[name]
	if !found {
		hook = &PathItem{}
		d.webHooks[name] = hook
	}

	if hook.Operations == nil {
		hook.Operations = make(map[string]*Operation, 3)
	} else if _, found := hook.Operations[method]; found {
		panic(fmt.Sprintf("已经存在 %s:%s 的 webhook", name, method))
	}
	hook.Operations[method] = o

	d.last = time.Now()
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
