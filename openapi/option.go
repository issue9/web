// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"fmt"

	"github.com/issue9/sliceutil"

	"github.com/issue9/web"
)

type Option func(*Document)

// WithResponse 指定所有接口都可能返回的对象类型
//
// 一般用于指定非正常状态的返回对象，比如 400 状态码的对象。
//
// status: 状态码；
// resp 返回对象，需要指定 resp.Ref.Ref；
//
// NOTE: 多次调用会依次添加
func WithResponse(status int, resp *Response) Option {
	return func(d *Document) {
		if resp.Ref == nil || resp.Ref.Ref == "" {
			panic("必须存在 ref")
		}
		resp.addComponents(d.components)

		if d.responses == nil {
			d.responses = make(map[int]string)
		}
		d.responses[status] = resp.Ref.Ref
	}
}

// WithMediaType 指定所有接口可用的媒体类型
//
// NOTE: 多次调用会依次添加，相同值会合并。
func WithMediaType(t ...string) Option {
	return func(d *Document) {
		if d.mediaTypes == nil {
			d.mediaTypes = t
		} else {
			d.mediaTypes = append(d.mediaTypes, t...)
			d.mediaTypes = sliceutil.Unique(d.mediaTypes, func(i, j string) bool { return i == j })
		}
	}
}

// WithHeader 指定所有请求都需要的报头
//
// 要求必须指定 p.Ref.Ref。
//
// NOTE: 多次调用会依次添加
func WithHeader(p ...*Parameter) Option {
	return func(d *Document) {
		for _, pp := range p {
			if pp.Ref == nil || pp.Ref.Ref == "" {
				panic("必须存在 ref")
			}
			pp.addComponents(d.components, InHeader)
			d.headers = append(d.headers, pp.Ref.Ref)
		}
	}
}

// WithHeader 指定所有请求都需要的 Cookie
//
// 要求必须指定 p.Ref.Ref。
//
// NOTE: 多次调用会依次添加
func WithCookie(p ...*Parameter) Option {
	return func(d *Document) {
		for _, pp := range p {
			if pp.Ref == nil || pp.Ref.Ref == "" {
				panic("必须存在 ref")
			}
			pp.addComponents(d.components, InCookie)
			d.cookies = append(d.cookies, pp.Ref.Ref)
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
	return func(d *Document) {
		d.info.license = newLicense(name, id)
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
		if d.servers == nil {
			d.servers = []*Server{}
		}

		d.servers = append(d.servers, &Server{
			URL:         url,
			Description: desc,
			Variables:   vars,
		})
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

// WithSecurity 指定 openapi.security 属性
//
// NOTE: 多次调用会依次添加
func WithSecurity(name string, key ...string) Option {
	return func(d *Document) {
		if d.security == nil {
			d.security = []*SecurityRequirement{}
		}
		d.security = append(d.security, &SecurityRequirement{Name: name, Values: key})
	}
}

// WithSecurityScheme 指定 openapi.components.securityScheme 属性
//
// NOTE: 多次调用会依次添加
func WithSecurityScheme(id string, s *SecurityScheme) Option {
	return func(d *Document) {
		if _, found := d.components.securitySchemes[id]; found {
			panic(fmt.Sprintf("已经存在名称为 %s 的项", id))
		}
		d.components.securitySchemes[id] = s
	}
}
