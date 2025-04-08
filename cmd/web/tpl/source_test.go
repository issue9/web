// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

package tpl

import (
	"go/token"
	"testing"

	"github.com/issue9/assert/v4"
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
	oldName := "o"
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
		out, err := replaceGoSourcePackageName(fset, item.file, []byte(item.in), oldName, newName)
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
