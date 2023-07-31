// SPDX-License-Identifier: MIT

// Package schema 将 ast 转换铖 openapi 的 schema 对象
package schema

import (
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

const refPrefix = "#/components/schemas/"

var refReplacer = strings.NewReplacer("/", ".")

type Ref = openapi3.SchemaRef

func NewRef(ref string, v *openapi3.Schema) *Ref {
	return openapi3.NewSchemaRef(ref, v)
}

// NewOpenAPI 声明基本的 openapi3.T 对象
func NewOpenAPI(ver string) *openapi3.T {
	c := openapi3.NewComponents()
	c.Schemas = make(openapi3.Schemas)
	c.Responses = make(openapi3.Responses)
	c.SecuritySchemes = make(openapi3.SecuritySchemes)

	return &openapi3.T{
		OpenAPI:    ver,
		Components: &c,
		Servers:    make(openapi3.Servers, 0, 5),
		Paths:      make(openapi3.Paths, 100),
		Security:   make(openapi3.SecurityRequirements, 0, 5),
		Tags:       make(openapi3.Tags, 0, 10),
	}
}

func addRefPrefix(ref *Ref) {
	if ref.Ref != "" && !strings.HasPrefix(ref.Ref, refPrefix) {
		ref.Ref = refPrefix + ref.Ref
	}
}
