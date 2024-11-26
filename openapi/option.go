// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"fmt"
	"strings"

	"github.com/issue9/web"
)

type Option func(*Document)

// WithHead 是否生成 HEAD 接口请求
func WithHead(enable bool) Option {
	return func(d *Document) { d.enableHead = enable }
}

// WithOptions 是否生成 OPTIONS 请求
func WithOptions(enable bool) Option {
	return func(d *Document) { d.enableOptions = enable }
}

// WithHTML 指定 HTML 模板
//
// 这将开启 [Document.Handler] 对 [html.Mimetype] 类型的支持。
// 此选项仅是定义了 [Document.Handler] 输出的一些内容，
// 还需要向 [github.com/issue9/web/mimetype/html] 添加模板。
//
// tpl 表示 HTML 模板名称，该模板由 [html.Install] 注册；
// assets 资源基地址，apidoc 的浏览工具一般都是在一个 html 文件中引入一些依赖文件，
// assets 就是这些依赖文件的基地址，可是指向一个 CDN 的地址，比如 swagger 的
// https://unpkg.com/swagger-ui-dist@5.18.2；
// logo LOGO 图标，可能还会被用于 favicon 等功能，请确保格式能被正确处理，可以为空；
//
// NOTE: [github.com/issue9/webuse/openapi] 下实现了部分简单的模板。
func WithHTML(tpl, assets, logo string) Option {
	return func(d *Document) {
		d.templateName = tpl
		d.assetsURL = assets
		d.logo = logo
	}
}

// WithResponse 向 components/responses 添加对象
//
// 一般用于指定非正常状态的返回对象，比如 400 状态码的对象。
//
// resp 返回对象，需要指定 resp.Ref.Ref，其它接口可以通过 Ref 引用该对象；
// status: 状态码，可以是 4XX 的形式，如果该值不为空，那么将当前对象以此状态码应用到所有的接口；
//
// NOTE: 多次调用会依次添加
func WithResponse(resp *Response, status ...string) Option {
	return func(d *Document) {
		if resp.Ref == nil || resp.Ref.Ref == "" {
			panic("resp 必须存在 ref")
		}
		resp.addToComponents(d.components)

		for _, s := range status {
			d.responses[s] = resp.Ref.Ref
		}
	}
}

// PresetOptions 提供 [web.Problem] 的 [Response] 对象并应用所有接口的 4XX 和 5XX 状态码
func WithProblemResponse() Option {
	return WithResponse(&Response{
		Ref:         &Ref{Ref: "problem"},
		Body:        NewSchema(web.Problem{}, web.Phrase("problem response schema"), web.Phrase("problem response schema desc")),
		Problem:     true,
		Description: web.Phrase("problem response"),
	}, "4XX", "5XX")
}

// WithMediaType 指定所有接口可用的媒体类型
//
// t 用于指定支持的媒体类型，必须是 [web.Server] 实例支持的类型。
//
// NOTE: 多次调用会依次添加，相同值会合并。
func WithMediaType(t ...string) Option {
	return func(d *Document) {
		for _, tt := range t {
			found := false

			for k, v := range d.s.Mimetypes() {
				if k == tt {
					found = true
					d.mediaTypes[tt] = v
				}
			}

			if !found {
				panic(fmt.Sprintf("不支持 %s 媒体类型", tt))
			}
		}
	}
}

// WithHeader 向 components/headers 添加对象
//
// p 要求必须指定 Ref.Ref，其它接口可以通过 Ref 引用该对象。
// global 是否将该对象默认应用在所有的接口上；
//
// NOTE: 多次调用会依次添加
func WithHeader(global bool, p ...*Parameter) Option {
	return func(d *Document) {
		for _, pp := range p {
			if pp.Ref == nil || pp.Ref.Ref == "" {
				panic("必须存在 ref")
			}

			if err := pp.valid(false); err != nil {
				panic(err)
			}

			pp.addToComponents(d.components, InHeader)

			if global {
				d.headers = append(d.headers, pp.Ref.Ref)
			}
		}
	}
}

// WithCookie 向 components/parameters 添加对象
//
// p 要求必须指定 Ref.Ref。其它接口可以通过 problem 引用此对象。
//
// global 是否将该对象默认应用在所有的接口上；
//
// NOTE: 多次调用会依次添加
func WithCookie(global bool, p ...*Parameter) Option {
	return func(d *Document) {
		for _, pp := range p {
			if pp.Ref == nil || pp.Ref.Ref == "" {
				panic("必须存在 ref")
			}
			pp.addToComponents(d.components, InCookie)
			if global {
				d.cookies = append(d.cookies, pp.Ref.Ref)
			}
		}
	}
}

// WithQuery 向 components/parameters 添加 InQuery 的对象
//
// p 要求必须指定 Ref.Ref。其它接口可以通过 ref 引用此对象。
//
// NOTE: 多次调用会依次添加
func WithQuery(p ...*Parameter) Option {
	return func(d *Document) {
		for _, pp := range p {
			if pp.Ref == nil || pp.Ref.Ref == "" {
				panic("必须存在 ref")
			}
			pp.addToComponents(d.components, InQuery)
		}
	}
}

// WithPath 向 components/parameters 添加 InPath 的对象
//
// p 要求必须指定 Ref.Ref。其它接口可以通过 ref 引用此对象。
//
// NOTE: 多次调用会依次添加
func WithPath(global bool, p ...*Parameter) Option {
	return func(d *Document) {
		for _, pp := range p {
			if pp.Ref == nil || pp.Ref.Ref == "" {
				panic("必须存在 ref")
			}
			pp.addToComponents(d.components, InPath)
		}
	}
}

// WithCallback 预定义回调对象
func WithCallback(c ...*Callback) Option {
	return func(d *Document) {
		for _, cc := range c {
			if cc.Ref == nil || cc.Ref.Ref == "" {
				panic("必须存在 ref")
			}
			if len(cc.Callback) == 0 {
				panic("Callback 不能为空")
			}

			cc.addToComponents(d.components)
		}
	}
}

// WithDescription 指定描述信息
//
// summary 指定的是 openapi.info.summary 属性；
// description 指定的是 openapi.info.description 属性；
//
// NOTE: 多次调用会相互覆盖
func WithDescription(summary, desc web.LocaleStringer) Option {
	return func(d *Document) {
		d.info.summary = summary
		d.info.description = desc
	}
}

// WithLicense 添加版权信息
//
// NOTE: 多次调用会相互覆盖
func WithLicense(name, id string) Option {
	if name == "" {
		panic("参数 name 不能为空")
	}

	return func(d *Document) {
		switch {
		case id == "":
			d.info.license = &licenseRenderer{Name: name}
		case strings.HasPrefix(id, "http://") || strings.HasPrefix(id, "https://"):
			d.info.license = &licenseRenderer{Name: name, URL: id}
		default:
			d.info.license = &licenseRenderer{Name: name, Identifier: id}
		}
	}
}

// WithContact 添加联系信息
//
// NOTE: 多次调用会相互覆盖
func WithContact(name, email, url string) Option {
	return func(d *Document) {
		d.info.contact = &contactRender{
			Name:  name,
			Email: email,
			URL:   url,
		}
	}
}

// WithTerms 服务条款的连接
//
// NOTE: 多次调用会相互覆盖
func WithTerms(url string) Option {
	return func(d *Document) {
		d.info.termsOfService = url
	}
}

// WithExternalDocs 指定扩展文档
//
// NOTE: 多次调用会相互覆盖
func WithExternalDocs(url string, desc web.LocaleStringer) Option {
	return func(d *Document) {
		d.externalDocs = &ExternalDocs{
			Description: desc,
			URL:         url,
		}
	}
}

// WithServer 添加 openapi.servers 变量
func WithServer(url string, desc web.LocaleStringer, vars ...*ServerVariable) Option {
	return func(d *Document) {
		s := &Server{
			URL:         url,
			Description: desc,
			Variables:   vars,
		}

		if err := s.valid(); err != nil {
			panic(err)
		}

		d.servers = append(d.servers, s)
	}
}

// WithTag 添加标签
//
// name 标签名称；
// desc 标签描述；
// extDocURL 扩展文档的地址；
// extDocDesc 扩展文档的描述；
//
// NOTE: 多次调用会依次添加
func WithTag(name string, desc web.LocaleStringer, extDocURL string, extDocDesc web.LocaleStringer) Option {
	return func(d *Document) {
		if d.tags == nil {
			d.tags = []*tag{}
		}

		t := &tag{
			name:        name,
			description: desc,
		}

		if extDocURL != "" {
			t.externalDocs = &ExternalDocs{
				Description: extDocDesc,
				URL:         extDocURL,
			}
		}

		d.tags = append(d.tags, t)
	}
}

// WithSecurityScheme 指定验证方案
//
// s 需要添加的验证方案；
// scope 如果指定了该值，那么会以 s.ID 为名称，scope 为值添加至 openapi.securiy，
// scope 如果是多个参数，每个参数应该都是不同的；
//
// NOTE: 多次调用会依次添加
func WithSecurityScheme(s *SecurityScheme, scope ...[]string) Option {
	return func(d *Document) {
		if _, found := d.components.securitySchemes[s.ID]; found {
			panic(fmt.Sprintf("已经存在名称为 %s 的项", s.ID))
		}

		if err := s.valid(); err != nil {
			panic(err)
		}

		d.components.securitySchemes[s.ID] = s

		if len(scope) > 0 {
			if d.security == nil {
				d.security = []*SecurityRequirement{}
			}

			for _, ss := range scope {
				d.security = append(d.security, &SecurityRequirement{Name: s.ID, Scopes: ss})
			}
		}
	}
}
