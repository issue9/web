// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"bytes"
	"reflect"
	"strconv"
	"strings"

	"github.com/issue9/errwrap"
	"golang.org/x/text/message"

	"github.com/issue9/web"
)

type parameterizedDoc struct {
	format string
	params []web.LocaleStringer
}

// ParameterizedDoc 声明带有参数的文档
//
// format 文档的内容，应该带有一个 %s 占位符，用于插入 params 参数。
// 如果已经存在相同的值，则是到该对象并插入 params，否则是声明新的对象。
func (d *Document) ParameterizedDoc(format string, params ...web.LocaleStringer) web.LocaleStringer {
	if !strings.Contains(format, "%s") {
		panic("参数 format 必须包含 '%s'")
	}

	if p, found := d.parameterizedDocs[format]; found {
		p.params = append(p.params, params...)
		return p
	} else {
		p := &parameterizedDoc{
			format: format,
			params: params,
		}
		d.parameterizedDocs[format] = p
		return p
	}
}

func (d *parameterizedDoc) LocaleString(p *message.Printer) string {
	buf := &bytes.Buffer{}
	for _, param := range d.params {
		buf.WriteString(param.LocaleString(p))
	}
	return p.Sprintf(d.format, buf)
}

// MarkdownProblems 将 problems 的内容生成为 markdown
//
// titleLevel 标题的级别，0-6：
// 如果取值为 0，表示以列表的形式输出，并忽略 detail 字段的内容。
// 1-6 表示输出 detail 内容，并且将 type 和 title 作为标题；
func MarkdownProblems(s web.Server, titleLevel int) web.LocaleStringer {
	if titleLevel != 0 {
		return markdownProblemsWithDetail(s, titleLevel)
	} else {
		return markdownProblemsWithoutDetail(s)
	}
}

func markdownProblemsWithoutDetail(s web.Server) web.LocaleStringer {
	buf := &errwrap.Buffer{}
	args := make([]any, 0, 30)
	for _, p := range s.Problems().Problems() {
		buf.Printf("- %s", p.Type()).WString(": %s\n\n")
		args = append(args, p.Title)
	}
	return web.Phrase(buf.String(), args...)
}

func markdownProblemsWithDetail(s web.Server, titleLevel int) web.LocaleStringer {
	ss := strings.Repeat("#", titleLevel)

	buf := &errwrap.Buffer{}
	args := make([]any, 0, 30)
	for _, p := range s.Problems().Problems() {
		buf.Printf("%s %s ", ss, p.Type()).
			WString("%s\n\n").
			WString("%s\n\n")
		args = append(args, p.Title, p.Detail)
	}

	return web.Phrase(buf.String(), args...)
}

// MarkdownGoObject 将 Go 对象转换为 markdown 表示方式
//
// 对于结构类型会自动展开。
//
// 如果 v 的字段包含 [CommentTag] 的标签内容，会尝试装将其本地化并作为注释放在字段之后。
//
// t 需要转换的类型；
// m 在此表中的类型会直接转换为键值表示的类型，而不是真实的类型。
func MarkdownGoObject(v any, m map[reflect.Type]string) web.LocaleStringer {
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	buf := &errwrap.Buffer{}
	params := goDefine(buf, make([]any, 0, 10), 0, t, m, false)

	s := buf.String()
	if strings.HasPrefix(s, "struct {") { // 结构可能由于 m 的关系返回一个非结构体的类型定义，所以只能由开头是否为 struct { 判断是否为结构体。
		s = "type " + t.Name() + " " + s
	}

	return web.Phrase("```go\n"+s+"\n```\n", params...)
}

func goDefine(buf *errwrap.Buffer, params []any, indent int, t reflect.Type, m map[reflect.Type]string, anonymous bool) []any {
	if len(m) > 0 {
		if s, found := m[t]; found {
			buf.WString(s)
			return params
		}
	}

	switch t.Kind() {
	case reflect.Func, reflect.Chan: // 忽略
	case reflect.Pointer:
		buf.WByte('*')
		return goDefine(buf, params, indent, t.Elem(), m, anonymous)
	case reflect.Slice:
		buf.WString("[]")
		return goDefine(buf, params, indent, t.Elem(), m, anonymous)
	case reflect.Array:
		buf.WByte('[').WString(strconv.Itoa(t.Len())).WByte(']')
		return goDefine(buf, params, indent, t.Elem(), m, anonymous)
	case reflect.Struct:
		if !anonymous {
			if t.NumField() == 0 {
				buf.WString("struct {}")
				return params
			}

			buf.WString("struct {\n")
			indent++
		}

		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)

			if f.Anonymous {
				tt := f.Type
				for tt.Kind() == reflect.Pointer { // 匿名字段需要去掉指针类型
					tt = tt.Elem()
				}
				params = goDefine(buf, params, indent, tt, m, true)
				continue
			}

			if !f.IsExported() {
				continue
			}

			if f.Type.Kind() == reflect.Func || f.Type.Kind() == reflect.Chan {
				continue
			}

			buf.WString(strings.Repeat("\t", indent)).WString(f.Name).WByte('\t')
			params = goDefine(buf, params, indent, f.Type, m, false)

			if f.Tag != "" {
				buf.WByte('\t').WByte('`').WString(string(f.Tag)).WByte('`')

				if c := f.Tag.Get(CommentTag); c != "" {
					buf.WString("\t// %s")
					params = append(params, web.Phrase(c))
				}
			}

			buf.WByte('\n')
		}

		if !anonymous {
			indent--
			buf.WString(strings.Repeat("\t", indent)).WByte('}')
		}
	default:
		buf.WString(t.Name())
	}

	return params
}
