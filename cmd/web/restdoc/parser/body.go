// SPDX-License-Identifier: MIT

package parser

import (
	"context"
	"errors"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/web"

	"github.com/issue9/web/cmd/web/restdoc/openapi"
	"github.com/issue9/web/cmd/web/restdoc/schema"
	"github.com/issue9/web/cmd/web/restdoc/utils"
)

const responsesRef = "#/components/responses/"

// @req * object.path *desc
// * 表示采用 [Parser.media]
func (p *Parser) parseRequest(ctx context.Context, o *openapi3.Operation, t *openapi.OpenAPI, suffix, filename, currPath string, ln int) {
	// NOTE: 目前无法为不同的 media type 指定不同的类型，如果要这样做，
	// 需要处理 components/schemas 不同 media types 具有相同名称的问题。

	words, l := utils.SplitSpaceN(suffix, 3)
	if l < 2 {
		p.syntaxError("@req", 2, filename, ln)
		return
	}

	s, err := p.schema.New(ctx, t, buildPath(currPath, words[1]), false)
	if err != nil {
		var serr *schema.Error
		if errors.As(err, &serr) {
			serr.Log(p.l, p.schema.Packages().FileSet())
			return
		}
		p.l.Error(err, filename, ln)
		return
	}

	r := openapi3.NewRequestBody().WithContent(p.newContents(s, strings.Split(words[0], ",")...)).WithDescription(words[2])
	o.RequestBody = &openapi3.RequestBodyRef{Value: r}
}

// @resp 200 text/* object.path *desc
func (p *Parser) parseResponse(ctx context.Context, resps map[string]*openapi3.Response, t *openapi.OpenAPI, suffix, filename, currPath string, ln int) (ok bool) {
	words, l := utils.SplitSpaceN(suffix, 4)
	if l < 3 {
		p.syntaxError("@resp", 3, filename, ln)
		return false
	}

	s, err := p.schema.New(ctx, t, buildPath(currPath, words[2]), false)
	if err != nil {
		var serr *schema.Error
		if errors.As(err, &serr) {
			serr.Log(p.l, p.schema.Packages().FileSet())
			return false
		}
		p.l.Error(err, filename, ln)
		return false
	}

	content := p.newContents(s, strings.Split(words[1], ",")...)
	if resp, found := resps[words[0]]; found {
		// NOTE: 按规定必须得有 description，但此处不作判断，可以用 "" 代替。
		resp.WithDescription(words[3]).WithContent(content)
	} else {
		resp = openapi3.NewResponse().WithDescription(words[3]).WithContent(content)
		resps[words[0]] = resp
	}

	return true
}

// @resp-header 200 header desc
func (p *Parser) parseResponseHeader(resps map[string]*openapi3.Response, suffix, filename string, ln int) bool {
	words, l := utils.SplitSpaceN(suffix, 3)
	if l != 3 {
		p.syntaxError("@resp-header", 3, filename, ln)
		return false
	}

	s := openapi3.NewSchemaRef("", openapi3.NewStringSchema())
	s.Value.Title = words[2]
	h := &openapi3.HeaderRef{Value: &openapi3.Header{Parameter: openapi3.Parameter{Schema: s}}}

	if resp, found := resps[words[0]]; found {
		if resp.Headers == nil {
			resp.Headers = make(openapi3.Headers, 5)
		}
		resp.Headers[words[1]] = h
	} else {
		resp := openapi3.NewResponse()
		resp.Headers = openapi3.Headers{words[1]: h}
		resps[words[0]] = resp
	}

	return true
}

// g 是否将定义为全局的对象也写入 o
func (p *Parser) addResponses(o *openapi3.Operation, resps map[string]*openapi3.Response, g bool) {
	if l := len(resps) + len(p.resps); o.Responses == nil && l > 0 {
		o.Responses = openapi3.NewResponsesWithCapacity(l)
	}

	if g { // 全局的定义在前，才会被本地定义覆盖。
		for key, resp := range p.resps {
			o.Responses.Set(key, &openapi3.ResponseRef{Ref: responsesRef + key, Value: resp})
		}
	}

	for status, r := range resps {
		if o.Responses.Value(status) != nil {
			p.l.Warning(web.Phrase("override global response %s", status))
		}
		o.Responses.Set(status, &openapi3.ResponseRef{Value: r})
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
