// SPDX-License-Identifier: MIT

package schema

import (
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// NewPath 根据 t 生成路径参数的 Schema
//
// 如果 t 的类型无法解析，则会尝试将其作为正则进行处理，如果还是不行则返回错误。
func NewPath(t string) (*Ref, error) {
	// NOTE: 都是基本类型，ref 都直接为空。

	switch strings.ToLower(t) {
	case "int", "integer":
		return openapi3.NewSchemaRef("", openapi3.NewInt64Schema()), nil
	case "bool", "boolean":
		return openapi3.NewSchemaRef("", openapi3.NewBoolSchema()), nil
	case "string", "str":
		return openapi3.NewSchemaRef("", openapi3.NewStringSchema()), nil
	case "number", "float", "float32", "float64":
		return openapi3.NewSchemaRef("", openapi3.NewFloat64Schema()), nil
	case "id":
		var id float64 = 1
		schema := openapi3.NewInt64Schema()
		schema.Min = &id
		return openapi3.NewSchemaRef("", schema), nil
	default:
		if _, err := regexp.Compile(t); err != nil {
			return nil, err
		}
		return openapi3.NewSchemaRef("", openapi3.NewSchema().WithPattern(t)), nil
	}
}
