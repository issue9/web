// SPDX-License-Identifier: MIT

package openapi

import (
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// AddAPI 添加一个 API
//
// 这会自动将路径参数向 PathItem 元素移动。
func (doc *OpenAPI) AddAPI(path, method string, o *openapi3.Operation) {
	doc.pathLocker.Lock()
	defer doc.pathLocker.Unlock()

	doc.doc.AddOperation(path, strings.ToUpper(method), o)

	if index := slices.IndexFunc(o.Parameters, isPathParams); index < 0 {
		return
	}

	p := doc.doc.Paths[path]
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
