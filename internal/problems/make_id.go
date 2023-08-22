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
	filename = "id.go"
	pkgName  = "problems"
)

func main() {
	buf := &errwrap.Buffer{}
	buf.WString(make.FileHeader).
		WString("package ").WString(pkgName).WString("\n\n").
		WString("import (\n").
		WString("\"net/http\"\n\n").
		WString("\"github.com/issue9/localeutil\"\n").
		WString(")\n\n")

	kvs, err := make.GetStatuses()
	if err != nil {
		panic(err)
	}

	makeID(buf, kvs)
	makeStatus(buf, kvs)
	makeIDs(buf, kvs)
	makeInitLocalesFunc(buf, kvs)

	if buf.Err != nil {
		panic(buf.Err)
	}

	if err = source.DumpGoSource(filename, buf.Bytes()); err != nil {
		panic(err)
	}
}

func makeID(buf *errwrap.Buffer, kvs []make.Pair) {
	buf.WString("const(\n")

	buf.WString(`ProblemAboutBlank = "about:blank"`).WString("\n\n")
	for _, item := range kvs {
		buf.Printf("%s=\"%s\"\n", make.ID(item), strconv.Itoa(item.Value))
	}

	buf.WString(")\n\n")
}

func makeStatus(buf *errwrap.Buffer, kvs []make.Pair) {
	buf.WString("var statuses=map[string]int{\n")

	for _, item := range kvs {
		buf.Printf("%s:%s,\n", make.ID(item), "http."+item.Name)
	}

	buf.WString("}\n\n")
}

func makeIDs(buf *errwrap.Buffer, kvs []make.Pair) {
	buf.WString("var ids=map[int]string{\n")

	for _, item := range kvs {
		buf.Printf("%s:%s,\n", "http."+item.Name, make.ID(item))
	}

	buf.WString("}\n\n")
}

func makeInitLocalesFunc(buf *errwrap.Buffer, kvs []make.Pair) {
	buf.WString("func(p*Problems)initLocales(){")

	for _, item := range kvs {
		id := make.ID(item)
		status := "http." + item.Name

		title := "problem." + strconv.Itoa(item.Value)
		detail := title + ".detail"
		title = "localeutil.StringPhrase(\"" + title + "\")"
		detail = "localeutil.StringPhrase(\"" + detail + "\")"

		buf.Printf(`p.Add(%s,%s,%s,%s)`, id, status, title, detail).WByte('\n')
	}
	buf.WString("}\n\n")
}
