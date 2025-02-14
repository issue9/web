// SPDX-FileCopyrightText: 2024-2025 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"cmp"
	"maps"
	"reflect"
	"slices"
	"strings"

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
