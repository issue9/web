// SPDX-License-Identifier: MIT

package parser

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/issue9/web/cmd/web/internal/restdoc/schema"
	"github.com/issue9/web/cmd/web/internal/restdoc/utils"
)

func (doc *Parser) parseAPI(t *openapi3.T, currPath, suffix string, lines []string, ln int, filename string) {
	opt := openapi3.NewOperation()
	opt.Responses = openapi3.NewResponses()

	words, l := utils.SplitSpaceN(suffix, 3) // GET /users *desc
	var method, path string
	if l < 2 {
		doc.l.Error(errSyntax, filename, ln)
		return
	}
	method, path = words[0], words[1]
	opt.Summary = words[2]

	var req request
	resps := map[string]*response{}

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		switch tag, suffix := utils.CutTag(line); strings.ToLower(tag) {
		case "@id": // @id get_users
			opt.OperationID = suffix
		case "@tag": // @tag t1 t2
			opt.Tags = strings.Fields(suffix)
		case "@header": // @header key *desc
			doc.addCookieHeader(opt, openapi3.ParameterInHeader, suffix, filename, ln+i)
		case "@cookie": // @cookie name *desc
			doc.addCookieHeader(opt, openapi3.ParameterInCookie, suffix, filename, ln+i)
		case "@path": // @path name type *desc
			doc.addPath(opt, suffix, filename, ln+i)
		case "@query": // @query object.path *desc
			doc.addQuery(t, opt, currPath, suffix, filename, ln+i)
		case "@req": // @req object.path *desc
			if !doc.parseRequest(&req, t, suffix, filename, currPath, ln+i) {
				return
			}
		case "@req-types": // @req-types application/json application/xml
			req.media = utils.SplitSpace(suffix)
		case "@resp": // @resp 200 object.path *desc
			if !doc.parseResponse(resps, t, suffix, filename, currPath, ln+i) {
				return
			}
		case "@resp-ref": // @resp-ref 200 name
			words, l := utils.SplitSpaceN(suffix, 2)
			if l != 2 {
				doc.l.Error(errSyntax, filename, ln+i)
				return
			}
			opt.Responses[words[0]] = &openapi3.ResponseRef{Ref: words[1]}
		case "@resp-types": // @resp-types status application/json application/xml
			if !doc.parseResponseType(resps, t, suffix, filename, currPath, ln+i) {
				return
			}

		case "resp-header": // @resp-header 200 h1 *desc
			if !doc.parseResponseHeader(resps, t, suffix, filename, currPath, ln+i) {
				return
			}
		case "##": // 可能是 ## callback
			// TODO
		default:
			opt.Description = strings.Join(lines[i:], " ")
		}
	}

	doc.addRequestBody(opt, &req)
	doc.addResponses(opt, resps)
	t.AddOperation(path, method, opt)
}

func (doc *Parser) addQuery(t *openapi3.T, opt *openapi3.Operation, currPath, suffix, filename string, ln int) {
	words, l := utils.SplitSpaceN(suffix, 2)
	if l < 1 {
		doc.l.Error(errSyntax, filename, ln)
		return
	}

	s, err := doc.search.New(t, currPath, words[0], true)
	if err != nil {
		if serr, ok := err.(*schema.Error); ok {
			serr.Log(doc.l, doc.fset)
			return
		}
		doc.l.Error(err, filename, ln)
		return
	}

	opt.AddParameter(&openapi3.Parameter{In: openapi3.ParameterInQuery, Schema: s, Description: words[1]})
}

func (doc *Parser) addPath(opt *openapi3.Operation, suffix, filename string, ln int) {
	words, l := utils.SplitSpaceN(suffix, 3)
	if l < 2 {
		doc.l.Error(errSyntax, filename, ln)
		return
	}

	s, err := schema.NewPath(words[1])
	if err != nil {
		doc.l.Error(err, filename, ln)
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
func (doc *Parser) addCookieHeader(opt *openapi3.Operation, in, suffix, filename string, ln int) {
	words, l := utils.SplitSpaceN(suffix, 2)
	if l < 1 {
		doc.l.Error(errSyntax, filename, ln)
		return
	}

	schema := openapi3.NewSchemaRef("", openapi3.NewStringSchema())
	opt.AddParameter(&openapi3.Parameter{In: in, Schema: schema, Name: words[0], Description: words[1]})
}
