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
	pkgName  = "server"
)

func main() {
	buf := &errwrap.Buffer{}
	buf.WString(make.FileHeader).
		WString("package ").WString(pkgName).WString("\n\n").
		WString("import (\n").
		WString("\"net/http\"\n\n").
		WString("\"github.com/issue9/web/internal/problems\"\n").
		WString("\"github.com/issue9/web\"\n").
		WString(")\n\n")

	kvs, err := make.GetStatuses()
	if err != nil {
		panic(err)
	}

	makeInitLocalesFunc(buf, kvs)

	if buf.Err != nil {
		panic(buf.Err)
	}

	if err = source.DumpGoSource(filename, buf.Bytes()); err != nil {
		panic(err)
	}
}

func makeInitLocalesFunc(buf *errwrap.Buffer, kvs []make.Pair) {
	buf.WString("func initProblems(p*problems.Problems){")

	for _, item := range kvs {
		status := "http." + item.Name
		title := "problem." + strconv.Itoa(item.Value)
		detail := title + ".detail"
		title = "web.StringPhrase(\"" + title + "\")"
		detail = "web.StringPhrase(\"" + detail + "\")"

		buf.Printf(`p.Add(web.%s,%s,%s,%s)`, item.ID(), status, title, detail).WByte('\n')
	}
	buf.WString("}\n\n")
}
