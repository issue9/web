// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

//go:build ignore

package main

import (
	"encoding/csv"
	"io/fs"
	"net/http"
	"strconv"
	"strings"

	"github.com/issue9/errwrap"
	"github.com/issue9/source/codegen"

	"github.com/issue9/web/internal/status"
)

const (
	pkgName  = "problems"
	filename = "generated_problems.go"
)

func main() {
	resp, err := http.Get("https://www.iana.org/assignments/http-problem-types/http-problem-type-uris.csv")
	if err != nil {
		panic(err)
	}
	ianaReader := csv.NewReader(resp.Body)
	ianaRecords, err := ianaReader.ReadAll()
	if err != nil {
		panic(err)
	}

	buf := &errwrap.Buffer{}
	buf.WString(status.FileHeader).
		WString("package ").WString(pkgName).WString("\n\n").
		WString("import (\n").
		WString(" \"net/http\"\n").
		WString(" \"unique\"\n\n").
		WString(" \"github.com/issue9/localeutil\"\n").
		WString(" )\n\n")

	kvs, err := status.Get()
	if err != nil {
		panic(err)
	}

	ianaSize := 0

	// 将 iana 中的 ID 加入到 kvs
	for index := 2; index < len(ianaRecords); index++ {
		ianaSize++
		record := ianaRecords[index]
		index := strings.LastIndexByte(record[0], '#')
		id := record[0][index+1:]

		kvs = append(kvs, status.Status{
			Name:    toCamel(id),
			Value:   id,
			Comment: record[3][1 : len(record[3])-1],
		})
	}

	makeID(buf, kvs, ianaSize)
	if err := makeInitProblems(buf, kvs, ianaRecords, ianaSize); err != nil {
		panic(err)
	}
	makeStatuses(buf, kvs, ianaSize)

	if buf.Err != nil {
		panic(buf.Err)
	}

	if err = codegen.Dump(filename, buf.Bytes(), fs.ModePerm); err != nil {
		panic(err)
	}
}

// 输出变量名以及对应的值
func makeID(buf *errwrap.Buffer, kvs []status.Status, ianaSize int) {
	buf.WString(`const AboutBlank = "about:blank"`).WString("\n\n")

	buf.WString("var(\n")
	buf.WString(`Empty = unique.Make("")`).WString("\n\n")

	for _, item := range kvs[:len(kvs)-ianaSize] {
		buf.Printf("%s=unique.Make(\"%s\") // %s\n", item.ID(), item.Value, item.Comment)
	}
	buf.WByte('\n')

	for _, item := range kvs[len(kvs)-ianaSize:] {
		buf.Printf("%s=unique.Make(\"%s\") // %s\n", item.Name, item.Value, item.Comment)
	}

	buf.WString(")\n\n")
}

// 生成 initProblems 方法
func makeInitProblems(buf *errwrap.Buffer, kvs []status.Status, ianaRecords [][]string, ianaSize int) error {
	buf.WString("func(p*Problems[T]) initProblems(){")

	// 标准状态码对应的 problem
	for _, item := range kvs[:len(kvs)-ianaSize] {
		status := "http." + item.Name

		s, err := strconv.Atoi(item.Value)
		if err != nil {
			return err
		}

		title := "localeutil.StringPhrase(\"" + http.StatusText(s) + "\")"

		buf.Printf(`p.add(&Item{Type:AboutBlank,ID:%s,Title:%s,Status:%s})`, item.ID(), title, status).WByte('\n')
	}

	buf.WByte('\n')

	// iana 的 problem
	for index := 2; index < len(ianaRecords); index++ {
		record := ianaRecords[index]

		index := strings.LastIndexByte(record[0], '#')
		id := record[0][index+1:]

		title := "localeutil.StringPhrase(\"" + record[1] + "\")"

		buf.Printf(`p.add(&Item{Type: "%s",ID:%s,Title:%s,Status:%s})`, record[0], toCamel(id), title, record[2]).WByte('\n')
	}

	buf.WString("}\n\n")

	return nil
}

// 生成 problemStatuses 变量
func makeStatuses(buf *errwrap.Buffer, kvs []status.Status, ianaSize int) {
	buf.WString("var problemStatuses=map[int]unique.Handle[string]{\n")

	for _, item := range kvs[:len(kvs)-ianaSize] {
		buf.Printf("%s:%s,\n", "http."+item.Name, item.ID())
	}

	buf.WString("}\n\n")
}

func toCamel(s string) string {
	parts := strings.Split(s, "-")
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, "")
}
