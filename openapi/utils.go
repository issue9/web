// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"cmp"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/issue9/errwrap"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"golang.org/x/text/message"

	"github.com/issue9/web"
)

var nameReplacer = strings.NewReplacer(
	"[", "_.",
	"]", "._",
	"/", ".",
)

func getTypeName(t reflect.Type) string {
	return nameReplacer.Replace(t.PkgPath() + "/" + t.Name())
}

// 可能返回 -，表示该字段不需要处理
// attr 表示是否 xml 的属性，仅针对 xml，其它类型无效。
func getTagName(field reflect.StructField, name string) (n string, omitempty, attr bool) {
	val := field.Tag.Get(name)
	if val == "-" {
		return "-", false, false
	}

	if val == "" {
		return field.Name, false, false
	}

	tags := strings.Split(val, ",")
	if len(tags) == 1 {
		return tags[0], false, false
	}
	return tags[0], slices.Index(tags[1:], "omitempty") >= 0, slices.Index(tags[1:], "attr") >= 0
}

func sprint(p *message.Printer, s web.LocaleStringer) string {
	if s == nil {
		return ""
	}
	return s.LocaleString(p)
}

// 将无序的 map 写入有序的 map 中
//
// 如果 m 为 nil，会初始化一个空对象。
//
// NOTE: 会对 in 的数据进行排序之后再输出。
func writeMap2OrderedMap[KEY cmp.Ordered, IN any, OUT any](in map[KEY]IN, m *orderedmap.OrderedMap[KEY, OUT], conv func(in IN) OUT) *orderedmap.OrderedMap[KEY, OUT] {
	if len(in) == 0 {
		return m
	}

	if m == nil {
		m = orderedmap.New[KEY, OUT](orderedmap.WithCapacity[KEY, OUT](len(in)))
	}

	keys := slices.Collect(maps.Keys(in))
	slices.Sort(keys)
	for _, k := range keys {
		m.Set(k, conv(in[k]))
	}

	return m
}

func getPathParams(path string) []string {
	ret := make([]string, 0, 3)

	for {
		if start := strings.IndexByte(path, '{'); start >= 0 {
			if end := strings.IndexByte(path[start:], '}'); end > 0 {
				ret = append(ret, path[start+1:start+end])
				path = path[start+end:]
				continue
			}
		}
		break
	}

	return ret
}

// MarkdownProblems 将 problems 的内容生成为 markdown
//
// titleLevel 标题的级别，0-6：
// 如果取值为 0，表示以列表的形式输出，并忽略 detail 字段的内容。
// 1-6 表示输出 detail 内容，并且将 type 和 title 作为标题；
func MarkdownProblems(s web.Server, titleLevel int, detail bool) web.LocaleStringer {
	if detail {
		return markdownProblemsWithDetail(s, titleLevel)
	} else {
		return markdownProblemsWithoutDetail(s)
	}
}

func markdownProblemsWithoutDetail(s web.Server) web.LocaleStringer {
	buf := &errwrap.Buffer{}
	args := make([]any, 0, 30)
	s.Problems().Visit(func(status int, lp *web.LocaleProblem) {
		buf.Printf("- %s", lp.Type()).WString(": %s\n\n")
		args = append(args, lp.Title)
	})
	return web.Phrase(buf.String(), args...)
}

func markdownProblemsWithDetail(s web.Server, titleLevel int) web.LocaleStringer {
	buf := &errwrap.Buffer{}
	ss := strings.Repeat("#", titleLevel)

	args := make([]any, 0, 30)
	s.Problems().Visit(func(status int, lp *web.LocaleProblem) {
		buf.Printf("%s %s ", ss, lp.Type()).
			WString("%s\n\n").
			WString("%s\n\n")
		args = append(args, lp.Title, lp.Detail)
	})

	return web.Phrase(buf.String(), args...)
}
