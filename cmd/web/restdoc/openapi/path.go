// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/mux/v9"
)

// AddAPI 添加一个 API
//
// 这会自动将路径参数向 PathItem 元素移动。
func (doc *OpenAPI) AddAPI(path string, o *openapi3.Operation, method string) {
	method = strings.ToUpper(method)

	var methods []string
	if method == "ANY" {
		methods = mux.AnyMethods()
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
