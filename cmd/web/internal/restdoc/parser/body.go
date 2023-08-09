// SPDX-License-Identifier: MIT

package parser

import (
	"errors"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/issue9/web/cmd/web/internal/restdoc/openapi"
	"github.com/issue9/web/cmd/web/internal/restdoc/schema"
	"github.com/issue9/web/cmd/web/internal/restdoc/utils"
)

type (
	response struct {
		schema *openapi3.SchemaRef
		desc   string
		media  []string
		header map[string]string
	}
)

// @req * object.path *desc
// * 表示采用 [Parser.media]
func (p *Parser) parseRequest(o *openapi3.Operation, t *openapi.OpenAPI, suffix, filename, currPath string, ln int) {
	// NOTE: 目前无法为不同的 media type 指定不同的类型，如果要这样做，
	// 需要处理 components/schemas 不同 media types 具有相同名称的问题。

	words, l := utils.SplitSpaceN(suffix, 3)
	if l < 2 {
		p.syntaxError("@req", 2, filename, ln)
		return
	}

	s, err := p.search.New(t, currPath, words[1], false)
	if err != nil {
		var serr *schema.Error
		if errors.As(err, &serr) {
			serr.Log(p.l, p.fset)
			return
		}
		p.l.Error(err, filename, ln)
		return
	}

	r := openapi3.NewRequestBody()
	r.Content = p.newContents(s, strings.Split(words[0], ",")...)
	r.Description = words[2]
	o.RequestBody = &openapi3.RequestBodyRef{Value: r}
}

// @resp 200 text/* object.path *desc
func (p *Parser) parseResponse(resps map[string]*response, t *openapi.OpenAPI, suffix, filename, currPath string, ln int) (ok bool) {
	words, l := utils.SplitSpaceN(suffix, 4)
	if l < 3 {
		p.syntaxError("@resp", 3, filename, ln)
		return false
	}

	s, err := p.search.New(t, currPath, words[2], false)
	if err != nil {
		var serr *schema.Error
		if errors.As(err, &serr) {
			serr.Log(p.l, p.fset)
			return false
		}
		p.l.Error(err, filename, ln)
		return false
	}

	if resp, found := resps[words[0]]; found {
		resp.desc = words[3]
		resp.schema = s
		resp.media = strings.Split(words[1], ",")
	} else {
		resps[words[0]] = &response{
			desc:   words[3],
			schema: s,
			media:  strings.Split(words[1], ","),
			header: make(map[string]string, 3),
		}
	}

	return true
}

// @resp-header 200 header desc
func (p *Parser) parseResponseHeader(resps map[string]*response, suffix, filename, currPath string, ln int) bool {
	words, l := utils.SplitSpaceN(suffix, 3)
	if l != 3 {
		p.syntaxError("@resp-header", 3, filename, ln)
		return false
	}

	if resp, found := resps[words[0]]; found {
		resp.header[words[1]] = words[2]
	} else {
		resps[words[0]] = &response{header: map[string]string{words[1]: words[2]}}
	}

	return true
}

func (p *Parser) addResponses(o *openapi3.Operation, resps map[string]*response) {
	for status, r := range resps {
		resp := openapi3.NewResponse()
		resp.Description = &r.desc
		resp.Content = p.newContents(r.schema, r.media...)

		resp.Headers = make(openapi3.Headers, len(r.header))
		for h, title := range r.header {
			s := openapi3.NewSchemaRef("", openapi3.NewStringSchema())
			s.Value.Title = title
			resp.Headers[h] = &openapi3.HeaderRef{
				Value: &openapi3.Header{Parameter: openapi3.Parameter{Schema: s}},
			}
		}

		if o.Responses == nil {
			o.Responses = openapi3.NewResponses()
		}
		o.Responses[status] = &openapi3.ResponseRef{Value: resp}
	}
}

// 当 media 为空时则直接采用 [Parser.media]
func (p *Parser) newContents(s *openapi3.SchemaRef, media ...string) openapi3.Content {
	if s == nil {
		return nil
	}

	c := openapi3.NewContent()

	if len(media) == 0 || len(media) == 1 && media[0] == "*" {
		media = p.media
	}

	mt := openapi3.NewMediaType()
	mt.Schema = s
	for _, m := range media {
		c[m] = mt
	}

	return c
}
