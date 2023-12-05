// SPDX-License-Identifier: MIT

package openapi

import (
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/assert/v3"
)

func TestNew(t *testing.T) {
	a := assert.New(t, false)

	doc := New("3.0")
	a.NotNil(doc).
		Equal(doc.Doc().OpenAPI, "3.0")
}

func TestOpenAPI_SaveAs(t *testing.T) {
	a := assert.New(t, false)

	doc := New("3.0")
	a.NotNil(doc)

	o := openapi3.NewOperation()
	o.Responses = openapi3.NewResponses()
	doc.AddAPI("/pet", o, http.MethodGet)

	a.NotError(doc.SaveAs("./openapi.out.yaml"))
	a.NotError(doc.SaveAs("./openapi.out.json"))
	a.ErrorString(doc.SaveAs("./openapi.out.xml"), "only support yaml and json")
}

func TestOpenAPI_Merge(t *testing.T) {
	a := assert.New(t, false)

	d2 := &openapi3.T{}
	d2.Tags = append(d2.Tags, &openapi3.Tag{Name: "t1"})
	d2.Paths = openapi3.NewPaths(openapi3.WithPath("/p1", &openapi3.PathItem{}))
	d2.Components = &openapi3.Components{
		Schemas: openapi3.Schemas{
			"Object": &openapi3.SchemaRef{Value: &openapi3.Schema{Type: "object"}},
		},
	}
	d2.Info = &openapi3.Info{Version: "1.0"}

	doc := New("3.0")
	a.NotNil(doc)
	doc.Merge(d2)
	a.Nil(doc.Doc().Info).
		Length(doc.Doc().Tags, 1).
		Length(doc.Doc().Components.Schemas, 1)
}
