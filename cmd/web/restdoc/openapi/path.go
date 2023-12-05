// SPDX-License-Identifier: MIT

package openapi

import (
	"net/http"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

const anyKey = "ANY"

var anyMethods = []string{
	http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPost, http.MethodPatch,
	http.MethodTrace, http.MethodDelete, http.MethodConnect, http.MethodOptions,
}

// AddAPI 添加一个 API
//
// 这会自动将路径参数向 PathItem 元素移动。
func (doc *OpenAPI) AddAPI(path string, o *openapi3.Operation, method string) {
	method = strings.ToUpper(method)

	var methods []string
	if method == anyKey {
		methods = anyMethods
	} else {
		methods = strings.Split(method, ",")
	}

	for _, m := range methods {
		doc.addAPI(path, m, o)
	}
}

func (doc *OpenAPI) addAPI(path, method string, o *openapi3.Operation) {
	doc.pathLocker.Lock()
	defer doc.pathLocker.Unlock()

	doc.doc.AddOperation(path, method, o)

	if index := slices.IndexFunc(o.Parameters, isPathParams); index < 0 {
		return
	}

	p := doc.doc.Paths.Find(path)
	if index := slices.IndexFunc(p.Parameters, isPathParams); index >= 0 {
		return
	}

	for _, param := range o.Parameters {
		if param.Value.In == openapi3.ParameterInPath {
			p.Parameters = append(p.Parameters, param)
		}
	}
	o.Parameters = slices.DeleteFunc(o.Parameters, isPathParams)
}

func isPathParams(p *openapi3.ParameterRef) bool { return p.Value.In == openapi3.ParameterInPath }
