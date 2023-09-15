// SPDX-License-Identifier: MIT

package parser

import (
	"go/build"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/version"
	"github.com/issue9/web"

	"github.com/issue9/web/cmd/web/git"
	"github.com/issue9/web/cmd/web/restdoc/openapi"
	"github.com/issue9/web/cmd/web/restdoc/utils"
)

// 解析 # restdoc 之后的内容
//
// title 表示 # restdoc 至该行的行尾内容；
// lines 表示第二行开始的所有内容，每一行不应该包含结尾的换行符；
// ln 表示 title 所在行的行号，在出错时，用于记录日志；
// filename 表示所在的文件，在出错时，用于记录日志；
func (p *Parser) parseRESTDoc(t *openapi.OpenAPI, currPath, title string, lines []string, ln int, filename string) {
	defer func() {
		if msg := recover(); msg != nil {
			p.l.Error(msg, filename, ln)
		}
	}()

	if t.Doc().Info != nil {
		p.l.Error(web.StringPhrase("dup # restdoc node"), filename, ln)
		return
	}

	info := &openapi3.Info{Title: title}
	p.resps = make(map[string]*openapi3.Response, 10)
	wait := make(map[int]string, 10) // @resp 依赖 @openapi 的内容，需要延后执行，行号：行内容。

	ln++ // lines 索引从 0 开始，所有行号需要加上 1 。
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
				p.syntaxError("@tag", 1, filename, ln+i)
				continue LOOP
			}

			if !p.isIgnoreTag(words[0]) {
				t.Doc().Tags = append(t.Doc().Tags, &openapi3.Tag{Name: words[0], Description: words[1]})
			}
		case "@server": // @server tag https://example.com *desc
			words, l := utils.SplitSpaceN(suffix, 3)
			if l < 2 {
				p.syntaxError("@server", 2, filename, ln+i)
				continue LOOP
			}
			tag := words[0]
			if tag == "*" {
				tag = ""
			}
			if tag != "" && p.isIgnoreTag(strings.Split(tag, ",")...) {
				continue
			}
			t.Doc().Servers = append(t.Doc().Servers, &openapi3.Server{URL: words[1], Description: words[2]})
		case "@license": // @license MIT *https://example.com/license
			words, l := utils.SplitSpaceN(suffix, 2)
			if l < 1 {
				p.syntaxError("@license", 1, filename, ln+i)
				continue LOOP
			}
			info.License = &openapi3.License{Name: words[0], URL: words[1]}
		case "@term": // @term https://example.com/term.html
			info.TermsOfService = suffix
		case "@version": // @version 1.0.0
			info.Version = p.parseVersion(suffix)
		case "@contact": // @contact name *https://example.com/contact *contact@example.com
			words, l := utils.SplitSpaceN(suffix, 3)
			if l == 0 {
				p.syntaxError("@contact", 1, filename, ln+i)
				continue LOOP
			}
			info.Contact = buildContact(words)
		case "@media": // @media application/json application/xml
			p.media = strings.Fields(suffix)
		case "@header": // @header h1 *desc
			words, l := utils.SplitSpaceN(suffix, 2)
			if l == 0 {
				p.syntaxError("@header", 1, filename, ln+i)
				continue LOOP
			}
			p.headers = append(p.headers, pair{key: words[0], desc: words[1]})
		case "@cookie": // @cookie c1 *desc
			words, l := utils.SplitSpaceN(suffix, 2)
			if l == 0 {
				p.syntaxError("@cookie", 1, filename, ln+i)
				continue LOOP
			}
			p.cookies = append(p.cookies, pair{key: words[0], desc: words[1]})
		case "@resp": // @resp 4XX text/* object.path desc
			// parseResponse 可能会引用 @openapi 的内容，需要在整个 LOOP 结束之后再执行。
			wait[ln+i] = suffix
		case "@resp-header": // @resp-header 4XX h1 *desc
			if !p.parseResponseHeader(p.resps, suffix, filename, currPath, ln+i) {
				continue LOOP
			}
		case "@scy-http": // @scy-http name scheme format *desc
			words, l := utils.SplitSpaceN(suffix, 4)
			if l < 3 {
				p.syntaxError("@scy-http", 3, filename, ln+i)
				continue LOOP
			}

			ss := openapi3.NewSecurityScheme()
			ss.Type = "http"
			ss.Scheme = words[1]
			ss.WithBearerFormat(words[2])
			ss.Description = words[3]
			t.Doc().Components.SecuritySchemes[words[0]] = &openapi3.SecuritySchemeRef{Value: ss}
		case "@scy-apikey": // @scy-apikey name param-name in *desc
			words, l := utils.SplitSpaceN(suffix, 4)
			if l < 3 {
				p.syntaxError("@scy-apikey", 3, filename, ln+i)
				continue LOOP
			}

			ss := openapi3.NewSecurityScheme()
			ss.Type = "apiKey"
			ss.Name = words[1]
			ss.In = words[2]
			ss.Description = words[3]
			t.Doc().Components.SecuritySchemes[words[0]] = &openapi3.SecuritySchemeRef{Value: ss}
		case "@scy-openid": // @scy-openid name url *desc
			words, l := utils.SplitSpaceN(suffix, 3)
			if l < 2 {
				p.syntaxError("@scy-openid", 2, filename, ln+i)
				continue LOOP
			}

			ss := openapi3.NewSecurityScheme()
			ss.Type = "openIdConnect"
			ss.OpenIdConnectUrl = words[1]
			ss.Description = words[2]
			t.Doc().Components.SecuritySchemes[words[0]] = &openapi3.SecuritySchemeRef{Value: ss}
		case "@scy-implicit": // @scy-implicit name authURL refreshURL scope1,scope2,...
			words, l := utils.SplitSpaceN(suffix, 4)
			if l < 3 {
				p.syntaxError("@scy-implicit", 3, filename, ln+i)
				continue LOOP
			}
			ss := openapi3.NewSecurityScheme()
			ss.Type = "oauth2"
			ss.Flows = &openapi3.OAuthFlows{
				Implicit: &openapi3.OAuthFlow{
					AuthorizationURL: words[1],
					RefreshURL:       words[2],
					Scopes:           parseScopes(words[3]),
				},
			}
			t.Doc().Components.SecuritySchemes[words[0]] = &openapi3.SecuritySchemeRef{Value: ss}
		case "@scy-password": // @scy-password name tokenURL refreshURL scope1,scope2,...
			words, l := utils.SplitSpaceN(suffix, 4)
			if l < 3 {
				p.syntaxError("@scy-password", 3, filename, ln+i)
				continue LOOP
			}
			ss := openapi3.NewSecurityScheme()
			ss.Type = "oauth2"
			ss.Flows = &openapi3.OAuthFlows{
				Password: &openapi3.OAuthFlow{
					TokenURL:   words[1],
					RefreshURL: words[2],
					Scopes:     parseScopes(words[3]),
				},
			}
			t.Doc().Components.SecuritySchemes[words[0]] = &openapi3.SecuritySchemeRef{Value: ss}
		case "@scy-code": // @scy-code name authURL tokenURL refreshURL scope1,scope2,....
			words, l := utils.SplitSpaceN(suffix, 5)
			if l < 4 {
				p.syntaxError("@scy-code", 4, filename, ln+i)
				continue LOOP
			}
			ss := openapi3.NewSecurityScheme()
			ss.Type = "oauth2"
			ss.Flows = &openapi3.OAuthFlows{
				AuthorizationCode: &openapi3.OAuthFlow{
					AuthorizationURL: words[1],
					TokenURL:         words[2],
					RefreshURL:       words[3],
					Scopes:           parseScopes(words[4]),
				},
			}
			t.Doc().Components.SecuritySchemes[words[0]] = &openapi3.SecuritySchemeRef{Value: ss}
		case "@scy-client": // @scy-client name tokenURL refreshURL scope1,scope2,...
			words, l := utils.SplitSpaceN(suffix, 4)
			if l < 3 {
				p.syntaxError("@scy-client", 3, filename, ln+i)
				continue LOOP
			}
			ss := openapi3.NewSecurityScheme()
			ss.Type = "oauth2"
			ss.Flows = &openapi3.OAuthFlows{
				ClientCredentials: &openapi3.OAuthFlow{
					TokenURL:   words[1],
					RefreshURL: words[2],
					Scopes:     parseScopes(words[3]),
				},
			}
			t.Doc().Components.SecuritySchemes[words[0]] = &openapi3.SecuritySchemeRef{Value: ss}
		case "@doc": // @doc url desc
			words, l := utils.SplitSpaceN(suffix, 2)
			if l < 1 {
				p.syntaxError("@doc", 1, filename, ln+i)
				continue LOOP
			}

			t.Doc().ExternalDocs = &openapi3.ExternalDocs{URL: words[0], Description: words[1]}
		case "@openapi":
			p.parseOpenAPI(t, suffix, filename, ln+i)
		default: // 不认识的标签，表示元数据部分结束，将剩余部分直接作为 info.Description
			if len(tag) > 1 && tag[0] == '@' {
				p.l.Warning(web.Phrase("unknown tag %s", tag))
			}
			info.Description = strings.Join(lines[i:], "\n")
			break LOOP
		}
	}

	for l, s := range wait {
		if !p.parseResponse(p.resps, t, s, filename, currPath, l) {
			return
		}
	}

	for status, r := range p.resps {
		t.Doc().Components.Responses[status] = &openapi3.ResponseRef{Value: r}
	}

	t.Doc().Info = info
}

func (p *Parser) parseVersion(suffix string) string {
	var hash string
	var err error

	switch suffix {
	case "git":
		hash, err = git.Commit(false)
	case "git-full":
		hash, err = git.Commit(true)
	default:
		return suffix
	}

	if err != nil { // 输出警告信息，但是不退出
		p.l.Warning(err)
	}

	ver, err := git.Version()
	if err != nil {
		p.l.Warning(err)
	}
	return ver + "+" + hash
}

func parseScopes(scope string) map[string]string {
	scopes := strings.Split(scope, ",")
	s := make(map[string]string, len(scopes))
	for _, ss := range scopes {
		s[ss] = ss
	}
	return s
}

// 引用另一个 openapi 文件
//
// 除了 info 之外的内容都将复制至 target。
// 包含以下几种格式：
//   - URI，支持 http、https 和 file。如果要指定相对路径的文件，需要省略 file:// 前缀；
//   - 相对于 Go 模块的文件地址，比如 github.com/issue9/cmfx@v0.1.1 restdoc.yaml
func (p *Parser) parseOpenAPI(target *openapi.OpenAPI, suffix, filename string, ln int) {
	words, l := utils.SplitSpaceN(suffix, 2)

	var uri *url.URL
	switch l {
	case 1:
		var err error
		if uri, err = url.Parse(words[0]); err != nil {
			p.l.Error(err, filename, ln)
			return
		}

		if uri.Scheme != "http" && uri.Scheme != "https" { // 除去 http 和 https 之外都默认为 file
			uri.Scheme = "" // 在 openapi3.Loader 中，scheme == "file" 时，可以省略。
		}
	case 2:
		modCache := filepath.Join(build.Default.GOPATH, "pkg", "mod")
		uri = &url.URL{Path: filepath.Join(modCache, words[0], words[1])}
	default:
		p.l.Error(web.StringPhrase("invalid format"), filename, ln)
		return
	}

	t, err := openapi3.NewLoader().LoadFromURI(uri)
	if err != nil {
		p.l.Error(err, filename, ln)
		return
	}

	ok, err := version.SemVerCompatible(t.OpenAPI, target.Doc().OpenAPI)
	if err != nil {
		p.l.Error(err, filename, ln)
		return
	}
	if !ok {
		p.l.Error(web.StringPhrase("version incompatible"), filename, ln)
		return
	}

	target.Merge(t)
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
