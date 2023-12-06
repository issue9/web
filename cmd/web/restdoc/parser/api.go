// SPDX-License-Identifier: MIT

package parser

import (
	"errors"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/web"

	"github.com/issue9/web/cmd/web/restdoc/openapi"
	"github.com/issue9/web/cmd/web/restdoc/schema"
	"github.com/issue9/web/cmd/web/restdoc/utils"
)

func (p *Parser) parseAPI(t *openapi.OpenAPI, currPath, suffix string, lines []string, ln int, filename string) {
	defer func() {
		// NOTE: recover 用于处理 openapi3 的 panic，但是不带行号信息。
		// 应当尽量在此之前查出错误。
		if msg := recover(); msg != nil {
			p.l.Error(msg, filename, ln)
		}
	}()

	words, l := utils.SplitSpaceN(suffix, 3) // GET /users *desc
	if l < 2 {
		p.syntaxError("# api", 2, filename, ln)
		return
	}

	method, path, summary := words[0], words[1], words[2]
	opt := openapi3.NewOperation()
	opt.Summary = summary

	resps := map[string]*openapi3.Response{}

	ln++ // lines 索引从 0 开始，所有行号需要加上 1 。
LOOP:
	for index := 0; index < len(lines); index++ {
		line := strings.TrimSpace(lines[index])
		if line == "" {
			continue
		}

		switch tag, suffix := utils.CutTag(line); strings.ToLower(tag) {
		case "@id": // @id get_users
			opt.OperationID = suffix
		case "@tag": // @tag t1 t2
			opt.Tags = strings.Fields(suffix)
			if p.isIgnoreTag(opt.Tags...) {
				p.l.Warning(web.Phrase("ignore %s", suffix))
				return
			}
		case "@header": // @header key *desc
			p.addCookieHeader("@header", opt, openapi3.ParameterInHeader, suffix, filename, ln+index)
		case "@cookie": // @cookie name *desc
			p.addCookieHeader("@cookie", opt, openapi3.ParameterInCookie, suffix, filename, ln+index)
		case "@path": // @path name type *desc
			p.addPath(opt, suffix, filename, ln+index)
		case "@query": // @query object.path
			p.addQuery(t, opt, currPath, suffix, filename, ln+index)
		case "@req": // @req text/* object.path *desc
			p.parseRequest(opt, t, suffix, filename, currPath, ln+index)
		case "@resp": // @resp 200 text/* object.path *desc
			if !p.parseResponse(resps, t, suffix, filename, currPath, ln+index) {
				return
			}
		case "@resp-header": // @resp-header 200 h1 *desc
			if !p.parseResponseHeader(resps, suffix, filename, currPath, ln+index) {
				return
			}
		case "##": // 可能是 ## callback
			delta := p.parseCallback(t, opt, currPath, suffix, lines[index:], ln+index, filename)
			index += delta
		default:
			if len(tag) > 1 && tag[0] == '@' {
				p.l.Warning(web.Phrase("unknown tag %s", tag))
			}
			opt.Description = strings.Join(lines[index:], " ")
			break LOOP
		}
	}

	for _, p := range p.headers {
		h := openapi3.NewHeaderParameter(p.key).WithDescription(p.desc).WithSchema(openapi3.NewStringSchema())
		opt.Parameters = append(opt.Parameters, &openapi3.ParameterRef{Value: h})
	}
	for _, p := range p.cookies {
		h := openapi3.NewCookieParameter(p.key).WithDescription(p.desc).WithSchema(openapi3.NewStringSchema())
		opt.Parameters = append(opt.Parameters, &openapi3.ParameterRef{Value: h})
	}
	p.addResponses(opt, resps, true)
	t.AddAPI(p.prefix+path, opt, method)
}

func (p *Parser) addQuery(t *openapi.OpenAPI, opt *openapi3.Operation, currPath, suffix, filename string, ln int) {
	if suffix == "" {
		p.syntaxError("@query", 1, filename, ln)
		return
	}

	s, err := p.schema.New(t, currPath, suffix, true)
	if err != nil {
		var serr *schema.Error
		if errors.As(err, &serr) {
			serr.Log(p.l, p.schema.Packages().FileSet())
			return
		}
		p.l.Error(err, filename, ln)
		return
	}

	if s.Value.Type != openapi3.TypeObject {
		p.l.Error(web.StringPhrase("@query must point to an object"), filename, ln)
		return
	}

	// 保证输出顺序相同，方便测试
	keys := make([]string, 0, len(s.Value.Properties))
	for name := range s.Value.Properties {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		v := s.Value.Properties[name]
		opt.AddParameter(&openapi3.Parameter{
			In:          openapi3.ParameterInQuery,
			Schema:      v,
			Description: v.Value.Description,
			Name:        name,
		})
	}
}

func (p *Parser) addPath(opt *openapi3.Operation, suffix, filename string, ln int) {
	words, l := utils.SplitSpaceN(suffix, 3)
	if l < 2 {
		p.syntaxError("@path", 2, filename, ln)
		return
	}

	s, err := schema.NewPath(words[1])
	if err != nil {
		p.l.Error(err, filename, ln)
		return
	}

	opt.AddParameter(&openapi3.Parameter{
		Schema:      s,
		In:          openapi3.ParameterInPath,
		Description: words[2],
		Name:        words[0],
		Required:    true,
	})
}

// 处理 @header 或是 @cookie
//
// 语法如下： @header h1 *desc 或是 @cookie c1 *desc
// 两者结构完全相同，处理方式也相同。
func (p *Parser) addCookieHeader(tag string, opt *openapi3.Operation, in, suffix, filename string, ln int) {
	words, l := utils.SplitSpaceN(suffix, 2)
	if l < 1 {
		p.syntaxError(tag, 1, filename, ln)
		return
	}

	s := openapi3.NewSchemaRef("", openapi3.NewStringSchema())
	opt.AddParameter(&openapi3.Parameter{In: in, Schema: s, Name: words[0], Description: words[1]})
}
