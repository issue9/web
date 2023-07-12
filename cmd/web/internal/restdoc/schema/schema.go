// SPDX-License-Identifier: MIT

// Package schema 将 ast 转换铖 openapi 的 schema 对象
package schema

import "github.com/getkin/kin-openapi/openapi3"

// NewOpenAPI 声明基本的 openapi3.T 对象
func NewOpenAPI() *openapi3.T {
	c := openapi3.NewComponents()
	c.Schemas = make(openapi3.Schemas)
	c.Responses = make(openapi3.Responses)
	c.SecuritySchemes = make(openapi3.SecuritySchemes)

	t := &openapi3.T{
		OpenAPI:    "3",
		Components: &c,
	}

	return t
}
