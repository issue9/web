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
	pkgName  = "server"
)

func main() {
	buf := &errwrap.Buffer{}
	buf.WString(status.FileHeader).
		WString("package ").WString(pkgName).WString("\n\n").
		WString("import (\n").
		WString("\"net/http\"\n\n").
		WString("\"github.com/issue9/web\"\n").
		WString(")\n\n")

	kvs, err := status.Get()
	if err != nil {
		panic(err)
	}

	makeInitLocalesFunc(buf, kvs)

	if buf.Err != nil {
		panic(buf.Err)
	}

	if err = codegen.Dump(filename, buf.Bytes(),fs.ModePerm); err != nil {
		panic(err)
	}
}

func makeInitLocalesFunc(buf *errwrap.Buffer, kvs []status.Pair) {
	buf.WString("func initProblems(p*problems){")

	for _, item := range kvs {
		status := "http." + item.Name
		title := "problem." + strconv.Itoa(item.Value)
		detail := title + ".detail"
		title = "web.StringPhrase(\"" + title + "\")"
		detail = "web.StringPhrase(\"" + detail + "\")"

		buf.Printf(`p.Add(%s,web.LocaleProblem{ID:web.%s,Title:%s,Detail:%s})`, status, item.ID(), title, detail).WByte('\n')
	}
	buf.WString("}\n\n")
}
