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
// format 描述信息的内容格式，应该带有一个 %s 占位符，用于插入 params 参数；
// f 用于将 params 转换为字符串，然后替换 format 中的 %s 占位符，如果为空，会自动将每个元素转换单独的一行；
//
// NOTE: 该对象只有作用于 [Operation.Description] 时，
// 才可通过 [Document.AppendDescriptionParameter] 向其追加内容的目的。
func ParameterizedDoc(format string, f func(string) string, params ...web.LocaleStringer) web.LocaleStringer {
	if !strings.Contains(format, "%s") {
		panic("参数 format 必须包含 %s")
	}

	if f == nil {
		f = func(s string) string { return s + "\n" }
	}

	return &parameterizedDoc{
		format: format,
		f:      f,
		params: params,
	}
}

func (d *parameterizedDoc) LocaleString(p *message.Printer) string {
	buf := &bytes.Buffer{}
	for _, param := range d.params {
		buf.WriteString(d.f(param.LocaleString(p)))
	}
	return p.Sprintf(d.format, buf)
}

// AppendDescriptionParameter 向指定的 API 文档的描述信息中添加信息
//
// 具体说明可参考 [ParameterizedDoc]
func (d *Document) AppendDescriptionParameter(operationID string, item ...web.LocaleStringer) {
	d.parameterizedDesc[operationID] = append(d.parameterizedDesc[operationID], item...)
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
