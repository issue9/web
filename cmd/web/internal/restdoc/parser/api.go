// SPDX-License-Identifier: MIT

package parser

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/sliceutil"
	"github.com/issue9/web"

	"github.com/issue9/web/cmd/web/internal/restdoc/schema"
	"github.com/issue9/web/cmd/web/internal/restdoc/utils"
)

func (p *Parser) parseAPI(t *openapi3.T, currPath, suffix string, lines []string, ln int, filename string, tags []string) {
	ln++ // lines 索引从 0 开始，所有行号需要加上 1 。

	defer func() {
		if msg := recover(); msg != nil {
			p.l.Error(msg, filename, ln)
		}
	}()

	ignore := func(tag ...string) bool {
		if len(tags) == 0 {
			return false
		}

		for _, t := range tag {
			if sliceutil.Exists(tags, func(tt string, _ int) bool { return tt == t }) {
				return false
			}
		}
		return true
	}

	words, l := utils.SplitSpaceN(suffix, 3) // GET /users *desc
	if l < 2 {
		p.l.Error(errSyntax, filename, ln)
		return
	}

	method, path, summary := words[0], words[1], words[2]
	opt := openapi3.NewOperation()
	opt.Responses = openapi3.NewResponses()
	opt.Summary = summary

	var req request
	resps := map[string]*response{}

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
			if ignore(opt.Tags...) {
				p.l.Warning(web.Phrase("ignore %s", suffix))
				return
			}
		case "@header": // @header key *desc
			p.addCookieHeader(opt, openapi3.ParameterInHeader, suffix, filename, ln+index)
		case "@cookie": // @cookie name *desc
			p.addCookieHeader(opt, openapi3.ParameterInCookie, suffix, filename, ln+index)
		case "@path": // @path name type *desc
			p.addPath(opt, suffix, filename, ln+index)
		case "@query": // @query name object.path *desc
			p.addQuery(t, opt, currPath, suffix, filename, ln+index)
		case "@req": // @req object.path *desc
			if !p.parseRequest(&req, t, suffix, filename, currPath, ln+index) {
				return
			}
		case "@req-media": // @req-media application/json application/xml
			req.media = strings.Fields(suffix)
		case "@resp": // @resp 200 object.path *desc
			if !p.parseResponse(resps, t, suffix, filename, currPath, ln+index) {
				return
			}
		case "@resp-ref": // @resp-ref 200 name
			words, l := utils.SplitSpaceN(suffix, 2)
			if l != 2 {
				p.l.Error(errSyntax, filename, ln+index)
				return
			}
			opt.Responses[words[0]] = &openapi3.ResponseRef{Ref: responsesRef + words[1]}
		case "@resp-media": // @resp-media status application/json application/xml
			if !p.parseResponseType(resps, t, suffix, filename, currPath, ln+index) {
				return
			}
		case "resp-header": // @resp-header 200 h1 *desc
			if !p.parseResponseHeader(resps, t, suffix, filename, currPath, ln+index) {
				return
			}
		case "##": // 可能是 ## callback
			// TODO
		default:
			opt.Description = strings.Join(lines[index:], " ")
			break
		}
	}

	p.addRequestBody(opt, &req)
	p.addResponses(opt, resps)
	t.AddOperation(path, method, opt)
}

func (p *Parser) addQuery(t *openapi3.T, opt *openapi3.Operation, currPath, suffix, filename string, ln int) {
	words, l := utils.SplitSpaceN(suffix, 3)
	if l < 2 {
		p.l.Error(errSyntax, filename, ln)
		return
	}

	s, err := p.search.New(t, currPath, words[1], true)
	if err != nil {
		if serr, ok := err.(*schema.Error); ok {
			serr.Log(p.l, p.fset)
			return
		}
		p.l.Error(err, filename, ln)
		return
	}

	opt.AddParameter(&openapi3.Parameter{
		In:          openapi3.ParameterInQuery,
		Schema:      s,
		Description: words[2],
		Name:        words[0],
	})
}

func (p *Parser) addPath(opt *openapi3.Operation, suffix, filename string, ln int) {
	words, l := utils.SplitSpaceN(suffix, 3)
	if l < 2 {
		p.l.Error(errSyntax, filename, ln)
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
	})
}

// 处理 @header 或是 @cookie
//
// 语法如下： @header h1 *desc 或是 @cookie c1 *desc
// 两者结构完全相同，处理方式也相同。
func (p *Parser) addCookieHeader(opt *openapi3.Operation, in, suffix, filename string, ln int) {
	words, l := utils.SplitSpaceN(suffix, 2)
	if l < 1 {
		p.l.Error(errSyntax, filename, ln)
		return
	}

	schema := openapi3.NewSchemaRef("", openapi3.NewStringSchema())
	opt.AddParameter(&openapi3.Parameter{In: in, Schema: schema, Name: words[0], Description: words[1]})
}
