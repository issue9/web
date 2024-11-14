// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"maps"
	"reflect"
	"slices"
	"strings"

	orderedmap "github.com/wk8/go-ordered-map/v2"
	"golang.org/x/text/message"

	"github.com/issue9/web"
)

var nameReplacer = strings.NewReplacer(
	"[", "__",
	"]", "__",
	"/", ".",
)

func getTypeName(t reflect.Type) string {
	return nameReplacer.Replace(t.PkgPath() + "/" + t.Name())
}

// 可能返回 -，表示该字段不需要处理
func getTagName(field reflect.StructField, name string) (n string, omitempty bool) {
	val := field.Tag.Get(name)
	if val == "-" {
		return "-", false
	}

	if val == "" {
		return field.Name, false
	}

	tags := strings.Split(val, ",")
	if len(tags) == 1 {
		return tags[0], false
	}
	return tags[0], slices.Index(tags[1:], "omitempty") >= 0
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
func writeMap2OrderedMap[IN any, OUT any](in map[string]IN, m *orderedmap.OrderedMap[string, OUT], conv func(in IN) OUT) *orderedmap.OrderedMap[string, OUT] {
	if len(in) == 0 {
		return m
	}

	if m == nil {
		m = orderedmap.New[string, OUT](orderedmap.WithCapacity[string, OUT](len(in)))
	}

	keys := slices.Collect(maps.Keys(in))
	slices.Sort(keys)
	for _, k := range keys {
		m.Set(k, conv(in[k]))
	}

	return m
}
