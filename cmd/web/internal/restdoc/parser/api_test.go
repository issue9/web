// SPDX-License-Identifier: MIT

package parser

import (
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/cmd/web/internal/restdoc/logger/loggertest"
	"github.com/issue9/web/cmd/web/internal/restdoc/schema"
)

func TestParser_parseAPI(t *testing.T) {
	a := assert.New(t, false)

	l := loggertest.New(a)
	p := New(l.Logger)
	doc := schema.NewOpenAPI("3")

	lines := []string{
		"@id post-admin",
		"@tag user",
		"@path id id id desc",
	}
	p.parseAPI(doc, "github.com/issue9/web", "POST /admins/{id}", lines, 5, "example.go", []string{"user"})
	path := doc.Paths["/admins/{id}"]
	a.NotNil(path).
		NotNil(path.Post).
		Nil(path.Delete).
		Length(path.Parameters, 1).
		Length(path.Post.Parameters, 0)

	lines = []string{
		"@id del-admin",
		"@tag user",
		"@path id id id desc",
	}
	p.parseAPI(doc, "github.com/issue9/web", "delete /admins/{id}", lines, 10, "example.go", []string{"user"})

	path = doc.Paths["/admins/{id}"]
	a.NotNil(path).
		NotNil(path.Delete).
		Length(path.Parameters, 1).
		Length(path.Delete.Parameters, 1) // 指定了 path 参数

	lines = []string{
		"@id put-admin",
		"@tag user",
	}
	p.parseAPI(doc, "github.com/issue9/web", "put /admins/{id}", lines, 15, "example.go", []string{"user"})

	path = doc.Paths["/admins/{id}"]
	a.NotNil(path).
		NotNil(path.Put).
		Length(path.Parameters, 1).
		Length(path.Put.Parameters, 0) // 未指定 path 参数，采用父元素
}
