// SPDX-FileCopyrightText: 2018-2025 caixw
//
// SPDX-License-Identifier: MIT

package htmldoc

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/web/cmd/web/locales"
)

func TestExport(t *testing.T) {
	a := assert.New(t, false)

	output := "./testdata/output.out.html"
	a.NotError(export("./testdata", "object", output, "zh-CN", "title", "desc", "", "")).
		FileExists(output)
}

func TestData_parse(t *testing.T) {
	a := assert.New(t, false)

	p, err := locales.NewPrinter("zh-CN")
	a.NotError(err).NotNil(p)

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, "./testdata", nil, parser.ParseComments)
	a.NotError(err).NotNil(pkgs)

	d := &data{}
	d.parse(p, "object", pkgs["testdata"].Files)
	a.Length(d.Objects, 2)

	o1 := d.Objects[0]
	a.Equal(o1.Title, "object").
		NotEmpty(o1.Desc).
		Length(o1.Items, 2)

	o2 := d.Objects[1]
	a.Equal(o2.Title, "obj2").
		NotEmpty(o2.Desc).
		Length(o2.Items, 1)
}

func TestGetName(t *testing.T) {
	a := assert.New(t, false)

	a.Equal(getName("name", "-"), "-").
		Equal(getName("name", "tag"), "tag").
		Equal(getName("name", ""), "name")
}
