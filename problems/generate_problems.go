// SPDX-FileCopyrightText: 2018-2026 caixw
//
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
	pkgName  = "problems"
	filename = "generated_problems.go"
)

func main() {
	buf := &errwrap.Buffer{}
	buf.WString(status.FileHeader).
		WString("package ").WString(pkgName).WString("\n\n").
		WString("import \"net/http\"\n\n").
		WString("import \"github.com/issue9/localeutil\"\n\n")

	kvs, err := status.Get()
	if err != nil {
		panic(err)
	}

	makeID(buf, kvs)
	makeInitProblems(buf, kvs)

	if buf.Err != nil {
		panic(buf.Err)
	}

	if err = codegen.Dump(filename, buf.Bytes(), fs.ModePerm); err != nil {
		panic(err)
	}
}

func makeID(buf *errwrap.Buffer, kvs []status.Pair) {
	buf.WString("const(\n")

	buf.WString(`AboutBlank = "about:blank"`).WString("\n\n")
	for _, item := range kvs {
		buf.Printf("%s=\"%s\"\n", item.ID(), strconv.Itoa(item.Value))
	}

	buf.WString(")\n\n")
}

func makeInitProblems(buf *errwrap.Buffer, kvs []status.Pair) {
	buf.WString("func(p*Problems[T]) initProblems(){")

	for _, item := range kvs {
		status := "http." + item.Name
		title := "problem." + strconv.Itoa(item.Value)
		title = "localeutil.StringPhrase(\"" + title + "\")"

		buf.Printf(`p.add(&Item{Type:"about:blank",ID:%s,Title:%s,Status:%s})`, item.ID(), title, status).WByte('\n')
	}
	buf.WString("}\n\n")
}
