// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

package tpl

import (
	"go/token"
	"path/filepath"
	"strings"
	"testing"

	"github.com/issue9/assert/v4"
	xcopy "github.com/otiai10/copy"
)

type testData struct {
	file string
	in   string
	out  string
	err  error
}

func TestReplaceGoSourcePackageName(t *testing.T) {
	a := assert.New(t, false)
	fset := token.NewFileSet()
	newName := "n"

	data := []*testData{
		{
			file: "f1.go",
			in: `package o
`,
			out: `package n
`,
		},

		{
			file: "f2_test.go",
			in: `package o_test
`,
			out: `package n_test
`,
		},
	}

	for _, item := range data {
		out, err := replaceGoSourcePackageName(fset, item.file, []byte(item.in), newName)
		if item.err != nil {
			a.ErrorIs(err, item.err, "not error at %s", item.file)
		} else {
			a.Equal(string(out), item.out, "not equal at %s", item.file)
		}
	}
}

func TestReplaceGoSourceImport(t *testing.T) {
	a := assert.New(t, false)
	fset := token.NewFileSet()
	oldPath := "example.com/old"
	newPath := "example.com/new"

	data := []*testData{
		{
			file: "f1.go",
			in: `package f1
import "example.com/old"
`,
			out: `package f1
import old "example.com/new"
`,
		},

		{
			file: "f2.go",
			in: `package f2
import (
	"go"
	"example.com/old"
)`,
			out: `package f2
import (
	"go"
	old "example.com/new"
)`,
		},

		{
			file: "f3.go",
			in: `package f3
import "example.com/old/sub"
`,
			out: `package f3
import "example.com/new/sub"
`,
		},

		{
			file: "f4.go",
			in: `package f4
import (
	"go"
	"example.com/old/sub"
)`,
			out: `package f4
import (
	"go"
	"example.com/new/sub"
)`,
		},
	}

	for _, item := range data {
		out, err := replaceGoSourceImport(fset, item.file, []byte(item.in), oldPath, newPath)
		if item.err != nil {
			a.ErrorIs(err, item.err, "not error at %s", item.file)
		} else {
			a.Equal(string(out), item.out, "not equal at %s", item.file)
		}
	}
}

func TestReplaceGo(t *testing.T) {
	a := assert.New(t, false)
	oldPath := "github.com/issue9/web/cmd/web/tpl/testdata/template"
	newPath := "github.com/issue9/web/cmd/web/tpl/testdata/out/goimport"
	dir := "./testdata/out/goimport"

	a.NotError(xcopy.Copy("./testdata/template", dir))
	a.NotError(replaceGo(dir, oldPath, newPath))

	// go.mod

	c := getFileContent(a, filepath.Join(dir, "go.mod"))
	a.True(strings.HasPrefix(c, "module "+newPath))

	// import 语句

	c = getFileContent(a, filepath.Join(dir, "sub/sub.go"))
	a.True(strings.Contains(c, `import template "github.com/issue9/web/cmd/web/tpl/testdata/out/goimport"`))

	// package name

	c = getFileContent(a, filepath.Join(dir, "template.go"))
	a.True(strings.HasPrefix(c, "package goimport"))
}
