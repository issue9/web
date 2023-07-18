// SPDX-License-Identifier: MIT

package parser

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/localeutil"

	"github.com/issue9/web/cmd/web/internal/restdoc/utils"
)

// 解析 # restdoc 之后的内容
//
// title 表示 # restdoc 至该行的行尾内容；
// lines 表示第二行开始的所有内容，每一行不应该包含结尾的换行符；
// ln 表示 title 所在行的行号，在出错时，用于记录日志；
// filename 表示所在的文件，在出错时，用于记录日志；
func (p *Parser) parseRESTDoc(t *openapi3.T, currPath, title string, lines []string, ln int, filename string) {
	ln++ // for lines 索引从 0 开始，所有行号需要加上 1 。

	info := &openapi3.Info{
		Title: title,
	}

	if t.Info != nil {
		p.l.Error(localeutil.Phrase("dup # restdoc note"), filename, ln)
		return
	}

	resps := make(map[string]*response, 10)

LOOP:
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		switch tag, suffix := utils.CutTag(line); strings.ToLower(tag) {
		case "@tag": // @tag name *desc
			words, l := utils.SplitSpaceN(suffix, 2)
			if l < 1 {
				p.l.Error(errSyntax, filename, ln+i)
				continue LOOP
			}
			t.Tags = append(t.Tags, &openapi3.Tag{Name: words[0], Description: words[1]})
		case "@server": // @server https://example.com *desc
			words, l := utils.SplitSpaceN(suffix, 2)
			if l < 1 {
				p.l.Error(errSyntax, filename, ln+i)
				continue LOOP
			}
			t.Servers = append(t.Servers, &openapi3.Server{URL: words[0], Description: words[1]})
		case "@license": // @license MIT *https://example.com/license
			words, l := utils.SplitSpaceN(suffix, 2)
			if l < 1 {
				p.l.Error(errSyntax, filename, ln+i)
				continue LOOP
			}
			info.License = &openapi3.License{Name: words[0], URL: words[1]}
		case "@term": // @term https://example.com/term.html
			info.TermsOfService = suffix
		case "@version": // @version 1.0.0
			info.Version = suffix
		case "@contact": // @contact name *https://example.com/contact *contact@example.com
			words, l := utils.SplitSpaceN(suffix, 3)
			if l == 0 {
				p.l.Error(errSyntax, filename, ln+i)
				continue LOOP
			}
			info.Contact = buildContact(words)
		case "@media": // @media application/json application/xml
			p.media = utils.SplitSpace(suffix)
		case "@resp": // @resp name object.path desc
			if !p.parseResponse(resps, t, suffix, filename, currPath, ln+i) {
				continue LOOP
			}
		case "@resp-types": // @resp-types name application/json application/xml
			if !p.parseResponseType(resps, t, suffix, filename, currPath, ln+i) {
				continue LOOP
			}
		case "resp-header": // @resp-header name h1 *desc
			if !p.parseResponseHeader(resps, t, suffix, filename, currPath, ln+i) {
				continue LOOP
			}
		case "@scy-http": // @scy-http name scheme format *desc
			words, l := utils.SplitSpaceN(suffix, 4)
			if l < 3 {
				p.l.Error(errSyntax, filename, ln+i)
				continue LOOP
			}

			ss := openapi3.NewSecurityScheme()
			ss.Type = "http"
			ss.Scheme = words[1]
			ss.WithBearerFormat(words[2])
			ss.Description = words[3]
			println(t.Components.SecuritySchemes == nil)
			t.Components.SecuritySchemes[words[0]] = &openapi3.SecuritySchemeRef{Value: ss}
		case "@scy-apikey": // @scy-apikey name param-name in *desc
			words, l := utils.SplitSpaceN(suffix, 4)
			if l < 3 {
				p.l.Error(errSyntax, filename, ln+i)
				continue LOOP
			}

			ss := openapi3.NewSecurityScheme()
			ss.Type = "apiKey"
			ss.Name = words[1]
			ss.In = words[2]
			ss.Description = words[3]
			t.Components.SecuritySchemes[words[0]] = &openapi3.SecuritySchemeRef{Value: ss}
		case "@scy-openid": // @scy-openid name url *desc
			words, l := utils.SplitSpaceN(suffix, 3)
			if l < 2 {
				p.l.Error(errSyntax, filename, ln+i)
				continue LOOP
			}

			ss := openapi3.NewSecurityScheme()
			ss.Type = "openIdConnect"
			ss.OpenIdConnectUrl = words[1]
			ss.Description = words[2]
			t.Components.SecuritySchemes[words[0]] = &openapi3.SecuritySchemeRef{Value: ss}
			// TODO 支持 security-oauth2 的相关功能
		case "@doc": // @doc url desc
			words, l := utils.SplitSpaceN(suffix, 2)
			if l < 1 {
				p.l.Error(errSyntax, filename, ln+i)
				continue LOOP
			}

			t.ExternalDocs = &openapi3.ExternalDocs{URL: words[0], Description: words[1]}
		default: // 不认识的标签，表示元数据部分结束，将剩余部分直接作为 info.Description
			info.Description = strings.Join(lines[i:], "\n")
			break LOOP
		}
	}

	for status, r := range resps {
		resp := openapi3.NewResponse()
		resp.Description = &r.desc
		resp.Content = p.newContents(r.schema, r.media...)
		t.Components.Responses[status] = &openapi3.ResponseRef{Value: resp}
	}

	t.Info = info
}

func buildContact(words []string) *openapi3.Contact {
	c := &openapi3.Contact{}
	for _, word := range words {
		switch {
		case utils.IsURL(word):
			c.URL = word
		case utils.IsEmail(word):
			c.Email = word
		default:
			c.Name = word
		}
	}

	return c
}
