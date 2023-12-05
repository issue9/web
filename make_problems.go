// SPDX-License-Identifier: MIT

//go:build ignore

package main

import (
	"io/fs"
	"strconv"

	"github.com/issue9/errwrap"
	"github.com/issue9/source/codegen"

	"github.com/issue9/web/internal/status"
)

const (
	filename = "problems.go"
	pkgName  = "web"
)

func main() {
	buf := &errwrap.Buffer{}
	buf.WString(status.FileHeader).
		WString("package ").WString(pkgName).WString("\n\n").
		WString("import \"net/http\"\n\n")

	kvs, err := status.Get()
	if err != nil {
		panic(err)
	}

	makeID(buf, kvs)
	makeIDs(buf, kvs)

	if buf.Err != nil {
		panic(buf.Err)
	}

	if err = codegen.Dump(filename, buf.Bytes(), fs.ModePerm); err != nil {
		panic(err)
	}
}

func makeID(buf *errwrap.Buffer, kvs []status.Pair) {
	buf.WString("const(\n")

	buf.WString(`ProblemAboutBlank = "about:blank"`).WString("\n\n")
	for _, item := range kvs {
		buf.Printf("%s=\"%s\"\n", item.ID(), strconv.Itoa(item.Value))
	}

	buf.WString(")\n\n")
}

func makeIDs(buf *errwrap.Buffer, kvs []status.Pair) {
	buf.WString("var problemsID=map[int]string{\n")

	for _, item := range kvs {
		buf.Printf("%s:%s,\n", "http."+item.Name, item.ID())
	}

	buf.WString("}\n\n")
}
