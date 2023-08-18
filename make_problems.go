// SPDX-License-Identifier: MIT

//go:build ignore

package main

import (
	"github.com/issue9/errwrap"
	"github.com/issue9/source"

	"github.com/issue9/web/internal/problems/make"
)

const (
	filename = "problems.go"
	pkgName  = "web"
)

func main() {
	buf := &errwrap.Buffer{}
	buf.WString(make.FileHeader).
		WString("package ").WString(pkgName).WString("\n\n").
		WString("import \"github.com/issue9/web/internal/problems\"\n\n")

	status, err := make.GetStatuses()
	if err != nil {
		panic(err)
	}

	buf.WString("const (\n")
	buf.WString(`ProblemAboutBlank = problems.ProblemAboutBlank`).WString("\n\n")
	for _, pair := range status {
		name := make.ID(pair)
		buf.Printf("%s=%s\n", name, "problems."+name)
	}
	buf.WString(")\n\n")

	if buf.Err != nil {
		panic(buf.Err)
	}

	if err := source.DumpGoSource(filename, buf.Bytes()); err != nil {
		panic(err)
	}
}
