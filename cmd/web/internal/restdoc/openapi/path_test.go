// SPDX-License-Identifier: MIT

package openapi

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/assert/v3"
)

func TestOpenAPI_AddAPI(t *testing.T) {
	a := assert.New(t, false)

	doc := New("3.0")
	a.NotNil(doc)

	o := openapi3.NewOperation()
	o.Parameters = append(o.Parameters, &openapi3.ParameterRef{Value: openapi3.NewPathParameter("p1")})
	doc.AddAPI("/path/{p1}", "GET", o)
	a.Length(doc.Doc().Paths["/path/{p1}"].Parameters, 1).
		Length(o.Parameters, 0) // 上移至 path item

	o = openapi3.NewOperation()
	o.Parameters = append(o.Parameters, &openapi3.ParameterRef{Value: openapi3.NewPathParameter("p1")})
	doc.AddAPI("/path/{p1}", "PUT", o)
	a.Length(doc.Doc().Paths["/path/{p1}"].Parameters, 1).
		Length(o.Parameters, 1) // path item 上已经有了，不再上移

	// 非路径参数
	o = openapi3.NewOperation()
	o.Parameters = append(o.Parameters, &openapi3.ParameterRef{Value: openapi3.NewQueryParameter("q1")})
	doc.AddAPI("/path/q1", "PUT", o)
	a.Length(doc.Doc().Paths["/path/q1"].Parameters, 0).
		Length(o.Parameters, 1)
}
