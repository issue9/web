// SPDX-License-Identifier: MIT

//go:build ignore

package main

import (
	"strconv"

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
		WString("import (\n").
		WString("\"net/http\"\n\n").
		WString("\"github.com/issue9/web/internal/problems\"\n").
		WString(")\n\n")

	kvs, err := make.GetStatuses()
	if err != nil {
		panic(err)
	}

	makeID(buf, kvs)
	makeIDs(buf, kvs)

	if buf.Err != nil {
		panic(buf.Err)
	}

	if err = source.DumpGoSource(filename, buf.Bytes()); err != nil {
		panic(err)
	}
}

func makeID(buf *errwrap.Buffer, kvs []make.Pair) {
	buf.WString("const(\n")

	buf.WString(`ProblemAboutBlank = problems.AboutBlank`).WString("\n\n")
	for _, item := range kvs {
		buf.Printf("%s=\"%s\"\n", item.ID(), strconv.Itoa(item.Value))
	}

	buf.WString(")\n\n")
}

func makeIDs(buf *errwrap.Buffer, kvs []make.Pair) {
	buf.WString("var problemsID=map[int]string{\n")

	for _, item := range kvs {
		buf.Printf("%s:%s,\n", "http."+item.Name, item.ID())
	}

	buf.WString("}\n\n")
}
