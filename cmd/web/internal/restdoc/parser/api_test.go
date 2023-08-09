// SPDX-License-Identifier: MIT

package parser

import (
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/cmd/web/internal/restdoc/logger/loggertest"
	"github.com/issue9/web/cmd/web/internal/restdoc/openapi"
)

func TestParser_parseAPI(t *testing.T) {
	a := assert.New(t, false)

	l := loggertest.New(a)
	p := New(l.Logger)
	doc := openapi.New("3")

	lines := []string{
		"@id post-admin",
		"@tag user",
		"@path id id id desc",
	}
	p.parseAPI(doc, "github.com/issue9/web", "POST /admins/{id}", lines, 5, "example.go", []string{"user"})
	path := doc.Doc().Paths["/admins/{id}"]
	a.NotNil(path).
		NotNil(path.Post).
		Nil(path.Delete).
		Length(path.Parameters, 1).
		Length(path.Post.Parameters, 0)
}
