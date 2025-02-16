// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"bytes"
	"strings"

	"github.com/issue9/errwrap"
	"golang.org/x/text/message"

	"github.com/issue9/web"
)

type parameterizedDoc struct {
	format string
	f      func(string) string
	params []web.LocaleStringer
}

// ParameterizedDoc 声明带有参数的文档
//
// format 文档的内容，应该带有一个 %s 占位符，用于插入 params 参数。
// 如果已经存在相同的值，则是到该对象并插入 params，否则是声明新的对象。
func (d *Document) ParameterizedDoc(format string, params ...web.LocaleStringer) web.LocaleStringer {
	if !strings.Contains(format, "%s") {
		panic("参数 format 必须包含 %s")
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
