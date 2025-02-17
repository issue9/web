// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"reflect"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/web"
)

func TestParameterizedDoc(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a)
	p := s.Locale().NewPrinter(language.SimplifiedChinese)
	d := New(s, web.Phrase("title"))

	a.PanicString(func() {
		d.ParameterizedDoc("format", nil)
	}, "参数 format 必须包含 '%s'")

	ls := d.ParameterizedDoc("format %s", web.Phrase("p1\n"))
	a.Length(ls.(*parameterizedDoc).params, 1).
		Length(d.parameterizedDocs, 1)

	ls = d.ParameterizedDoc("format %s", web.Phrase("p2\n"))
	a.Length(ls.(*parameterizedDoc).params, 2).
		Equal(ls.LocaleString(p), web.Phrase("format %s", "p1\np2\n").LocaleString(p))
}

type object1 struct {
	Int   int `json:"int" yaml:"int"`
	Array [5]int
	Slice []string
	Byte  byte
	Chan  chan int
}

type object2 struct {
	Int    int      `json:"int" yaml:"int"`
	Object *object1 `json:"object"`
}

type object3 struct {
	int int
	T   time.Time `json:"t"`
}

type object4 struct {
	*object3
	Int int
}

type object5 struct {
	*object4
	Str string `comment:"C"`
}

type object6 struct {
	XMLName struct{} `json:"root"`
	Str     []*object3
}

func TestMarkdownGoObject(t *testing.T) {
	a := assert.New(t, false)
	p := message.NewPrinter(language.SimplifiedChinese)

	wont := "```go\ntype object1 struct {\n\tInt\tint\t`json:\"int\" yaml:\"int\"`\n\tArray\t[5]int\n\tSlice\t[]string\n\tByte\tuint8\n}\n```\n"
	a.Equal(MarkdownGoObject(object1{}, nil).LocaleString(p), wont)

	wont = "```go\ntype object2 struct {\n\tInt\tint\t`json:\"int\" yaml:\"int\"`\n\tObject\t*struct {\n\t\tInt\tint\t`json:\"int\" yaml:\"int\"`\n\t\tArray\t[5]int\n\t\tSlice\t[]string\n\t\tByte\tuint8\n\t}\t`json:\"object\"`\n}\n```\n"
	a.Equal(MarkdownGoObject(&object2{}, nil).LocaleString(p), wont)

	a.Equal(MarkdownGoObject(5, nil).LocaleString(p), "```go\nint\n```\n")

	a.Equal(MarkdownGoObject("", nil).LocaleString(p), "```go\nstring\n```\n")

	a.Equal(MarkdownGoObject(func() {}, nil).LocaleString(p), "```go\n\n```\n")

	m := map[reflect.Type]string{reflect.TypeFor[time.Time](): "string"}

	a.Equal(MarkdownGoObject(time.Time{}, m).LocaleString(p), "```go\nstring\n```\n")
	a.Equal(MarkdownGoObject(&time.Time{}, m).LocaleString(p), "```go\nstring\n```\n")
	a.Equal(MarkdownGoObject(time.Time{}, nil).LocaleString(p), "```go\ntype Time struct {\n}\n```\n")

	a.Equal(MarkdownGoObject(&object3{}, m).LocaleString(p), "```go\ntype object3 struct {\n\tT\tstring\t`json:\"t\"`\n}\n```\n")
	a.Equal(MarkdownGoObject(&object4{}, m).LocaleString(p), "```go\ntype object4 struct {\n\tT\tstring\t`json:\"t\"`\n\tInt\tint\n}\n```\n")
	a.Equal(MarkdownGoObject(&object5{}, m).LocaleString(p), "```go\ntype object5 struct {\n\tT\tstring\t`json:\"t\"`\n\tInt\tint\n\tStr\tstring\t`comment:\"C\"`\t// C\n}\n```\n")
	a.Equal(MarkdownGoObject(object6{}, m).LocaleString(p), "```go\ntype object6 struct {\n\tXMLName\tstruct {}\t`json:\"root\"`\n\tStr\t[]*struct {\n\t\tT\tstring\t`json:\"t\"`\n\t}\n}\n```\n")
}
