// SPDX-License-Identifier: MIT

package parser

import (
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/issue9/web"
	"github.com/issue9/web/cmd/web/internal/restdoc/utils"
)

func (p *Parser) parseCallback(t *openapi3.T, o *openapi3.Operation, currPath, suffix string, lines []string, ln int, filename string) (delta int) {
	// callback name post $request.url *desc
	words, l := utils.SplitSpaceN(suffix, 5)
	if l < 4 {
		p.syntaxError("# callback", 4, filename, ln)
		return 0
	}

	if strings.ToLower(words[0]) != "callback" {
		return 0
	}

	name, method, path, summary := words[1], words[2], words[3], words[4]
	opt := openapi3.NewOperation()
	opt.Responses = openapi3.NewResponses()
	opt.Summary = summary

	resps := map[string]*response{}

	for delta = 0; delta < len(lines); delta++ {
		line := strings.TrimSpace(lines[delta])
		if line == "" {
			continue
		}

		switch tag, suffix := utils.CutTag(line); strings.ToLower(tag) {
		case "@header": // @header key *desc
			p.addCookieHeader(opt, openapi3.ParameterInHeader, suffix, filename, ln+delta)
		case "@cookie": // @cookie name *desc
			p.addCookieHeader(opt, openapi3.ParameterInCookie, suffix, filename, ln+delta)
		case "@query": // @query object.path *desc
			p.addQuery(t, opt, currPath, suffix, filename, ln+delta)
		case "@req": // @req text/* object.path *desc
			p.parseRequest(opt, t, suffix, filename, currPath, ln+delta)
		case "@resp": // @resp 200 text/* object.path *desc
			if !p.parseResponse(resps, t, suffix, filename, currPath, ln+delta) {
				return delta
			}
		case "@resp-ref": // @resp-ref 200 name
			words, l := utils.SplitSpaceN(suffix, 2)
			if l != 2 {
				p.syntaxError("@resp-ref", 2, filename, ln+delta)
				return delta
			}
			opt.Responses[words[0]] = &openapi3.ResponseRef{Ref: responsesRef + words[1]}
		case "@resp-header": // @resp-header 200 h1 *desc
			if !p.parseResponseHeader(resps, t, suffix, filename, currPath, ln+delta) {
				return delta
			}
		}
	}

	pi := &openapi3.PathItem{}
	switch method {
	case http.MethodConnect:
		pi.Connect = opt
	case http.MethodDelete:
		pi.Delete = opt
	case http.MethodGet:
		pi.Get = opt
	case http.MethodHead:
		pi.Head = opt
	case http.MethodOptions:
		pi.Options = opt
	case http.MethodPatch:
		pi.Patch = opt
	case http.MethodPost:
		pi.Post = opt
	case http.MethodPut:
		pi.Put = opt
	case http.MethodTrace:
		pi.Trace = opt
	default:
		p.l.Error(web.NewLocaleError("invalid http method %s", method), filename, ln)
		return 0
	}

	if o.Callbacks == nil {
		o.Callbacks = make(openapi3.Callbacks, 5)
	}
	callback := openapi3.Callback{path: pi}
	o.Callbacks[name] = &openapi3.CallbackRef{
		Value: &callback,
	}

	return delta
}
