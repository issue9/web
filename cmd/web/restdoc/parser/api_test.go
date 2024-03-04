// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package parser

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/issue9/assert/v4"

	"github.com/issue9/web/cmd/web/restdoc/logger/loggertest"
	"github.com/issue9/web/cmd/web/restdoc/openapi"
)

func TestParser_parseAPI(t *testing.T) {
	a := assert.New(t, false)

	l := loggertest.New(a)
	p := New(l.Logger, "", []string{"user"})
	doc := openapi.New("3")

	lines := []string{
		"@id post-admin",
		"@tag user",
		"@path id id id desc",
	}
	p.parseAPI(context.Background(), doc, "github.com/issue9/web", "POST /admins/{id}", lines, 5, "example.go")
	path := doc.Doc().Paths.Find("/admins/{id}")
	a.NotNil(path).
		NotNil(path.Post).
		Nil(path.Delete).
		Length(path.Parameters, 1).
		Length(path.Post.Parameters, 0)
}

func TestParser_parseSecurity(t *testing.T) {
	a := assert.New(t, false)

	l := loggertest.New(a)
	p := New(l.Logger, "", []string{"user"})

	opt := openapi3.NewOperation()
	a.Nil(opt.Security)
	p.parseSecurity(opt, "")
	a.NotNil(opt.Security).Empty((*opt.Security)[0])

	opt = openapi3.NewOperation()
	p.parseSecurity(opt, "oauth")
	a.Length(opt.Security, 1).Empty((*opt.Security)[0]["oauth"])

	opt = openapi3.NewOperation()
	p.parseSecurity(opt, "oauth arg1 arg2")
	a.Length(opt.Security, 1).
		Length((*opt.Security)[0]["oauth"], 2, "%+v", (*opt.Security)[0])
}
