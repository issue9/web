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

	// 理论上同一路由的路径参数应该是相同的，提取第一个添加的 Operation 路径参数作为整个 Path 的参数说明。

	// NOTE: 因为无法确保调用 addAPI 的调用顺序，所以如果每个 Operation 对路径参数的描述都是不同的，
	// 那么在生成的文档中，PathItem.Parameters 可能也是不同的。

	if index := slices.IndexFunc(o.Parameters, isPathParams); index < 0 { // 未指定路径参数
		return
	}

	pathItem := doc.doc.Paths.Find(path)
	if index := slices.IndexFunc(pathItem.Parameters, isPathParams); index >= 0 { // 已由其它 Operation 添加了参数
		return
	}

	for _, param := range o.Parameters {
		if param.Value.In == openapi3.ParameterInPath {
			pathItem.Parameters = append(pathItem.Parameters, param)
		}
	}
}

func isPathParams(p *openapi3.ParameterRef) bool { return p.Value.In == openapi3.ParameterInPath }
