// SPDX-License-Identifier: MIT

package parser

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/issue9/web/cmd/web/internal/restdoc/schema"
	"github.com/issue9/web/cmd/web/internal/restdoc/utils"
)

type (
	request struct {
		schema *openapi3.SchemaRef
		desc   string
		media  []string
	}

	response struct {
		schema *openapi3.SchemaRef
		desc   string
		media  []string
		header map[string]string
	}
)

// 解析 @req 的内容并将其写入 req
func (p *Parser) parseRequest(req *request, t *openapi3.T, suffix, filename, currPath string, ln int) (ok bool) {
	words, l := utils.SplitSpaceN(suffix, 2)
	if l < 1 {
		p.l.Error(errSyntax, filename, ln)
		return false
	}

	s, err := p.search.New(t, currPath, words[0], false)
	if err != nil {
		if serr, ok := err.(*schema.Error); ok {
			serr.Log(p.l, p.fset)
			return false
		}
		p.l.Error(err, filename, ln)
		return false
	}

	req.schema = s
	req.desc = words[1]

	return true
}

func (p *Parser) addRequestBody(o *openapi3.Operation, r *request) {
	req := openapi3.NewRequestBody()
	req.Content = p.newContents(r.schema, r.media...)
	req.Description = r.desc
	o.RequestBody = &openapi3.RequestBodyRef{Value: req}
}

// 解析 @resp 内容至 resps
func (p *Parser) parseResponse(resps map[string]*response, t *openapi3.T, suffix, filename, currPath string, ln int) (ok bool) {
	words, l := utils.SplitSpaceN(suffix, 3)
	if l < 2 {
		p.l.Error(errSyntax, filename, ln)
		return false
	}

	s, err := p.search.New(t, currPath, words[1], false)
	if err != nil {
		if serr, ok := err.(*schema.Error); ok {
			serr.Log(p.l, p.fset)
			return false
		}
		p.l.Error(err, filename, ln)
		return false
	}

	if resp, found := resps[words[0]]; found {
		resp.desc = words[2]
		resp.schema = s
	}
	resps[words[0]] = &response{desc: words[2], schema: s}

	return true
}

// @resp-header 200 header desc
func (p *Parser) parseResponseHeader(resps map[string]*response, t *openapi3.T, suffix, filename, currPath string, ln int) bool {
	words, l := utils.SplitSpaceN(suffix, 3)
	if l != 3 {
		p.l.Error(errSyntax, filename, ln)
		return false
	}

	if resp, found := resps[words[0]]; found {
		resp.header[words[1]] = words[2]
	}
	resps[words[0]] = &response{header: map[string]string{words[1]: words[2]}}

	return true
}

// @resp-type 200 application/json application/xml
func (p *Parser) parseResponseType(resps map[string]*response, t *openapi3.T, suffix, filename, currPath string, ln int) bool {
	words, l := utils.SplitSpaceN(suffix, 2)
	if l != 2 {
		p.l.Error(errSyntax, filename, ln)
		return false
	}

	types := strings.Fields(words[1])
	if resp, found := resps[words[0]]; found {
		resp.media = append(resp.media, types...)
	}
	resps[words[0]] = &response{media: types}

	return true
}

func (p *Parser) addResponses(o *openapi3.Operation, resps map[string]*response) {
	for status, r := range resps {
		resp := openapi3.NewResponse()
		resp.Description = &r.desc
		resp.Content = p.newContents(r.schema, r.media...)
		resp.Headers = make(openapi3.Headers, len(r.header))
		for h, desc := range r.header {
			schema := openapi3.NewSchemaRef("", openapi3.NewStringSchema())
			p := openapi3.Parameter{In: openapi3.ParameterInHeader, Schema: schema, Description: desc}
			resp.Headers[h] = &openapi3.HeaderRef{Value: &openapi3.Header{Parameter: p}}
		}

		if o.Responses == nil {
			o.Responses = openapi3.NewResponses()
		}
		o.Responses[status] = &openapi3.ResponseRef{Value: resp}
	}
}

// 当 media 为空时则直接采用 doc.media
func (p *Parser) newContents(s *openapi3.SchemaRef, media ...string) openapi3.Content {
	c := openapi3.NewContent()

	if len(media) == 0 {
		media = p.media
	}

	mt := openapi3.NewMediaType()
	mt.Schema = s
	for _, m := range media {
		c[m] = mt
	}

	return c
}
