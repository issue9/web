// SPDX-License-Identifier: MIT

//go:build ignore

package main

import (
	"io/fs"

	"github.com/issue9/errwrap"
	"github.com/issue9/source/codegen"

	"github.com/issue9/web/internal/status"
)

const (
	filename = "statuses.go"
	pkgName  = "status"
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

	makeStatuses(buf, kvs)

	if buf.Err != nil {
		panic(buf.Err)
	}

	if err = codegen.Dump(filename, buf.Bytes(), fs.ModePerm); err != nil {
		panic(err)
	}
}

func makeStatuses(buf *errwrap.Buffer, kvs []status.Pair) {
	buf.WString("var problemStatuses=map[int]struct{}{\n")

	for _, item := range kvs {
		buf.Printf("%s:{},\n", "http."+item.Name)
	}

	buf.WString("}\n\n")
}
