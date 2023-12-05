// SPDX-License-Identifier: MIT

package enum

import "strings"

type data struct {
	FileHeader string
	Package    string
	Enums      []*enum
}

// enum 枚举类型的数据
type enum struct {
	Name     string  // 类型名称
	Values   []value // 类型的所有可能值
	Receiver string

	Type2StringMap string
	String2TypeMap string
}

type value struct {
	Name   string // 值名称
	String string // 值对应的字符串值
}

func newEnum(t string, vals ...string) *enum {
	has := true
	for _, v := range vals {
		if has = has && strings.HasPrefix(v, t); !has {
			break
		}
	}

	values := make([]value, 0, len(vals))
	if has {
		for _, v := range vals {
			values = append(values, value{Name: v, String: strings.ToLower(strings.TrimPrefix(v, t))})
		}
	} else {
		for _, v := range vals {
			values = append(values, value{Name: v, String: strings.ToLower(v)})
		}
	}

	return &enum{
		Name:     t,
		Values:   values,
		Receiver: string(t[0]),

		Type2StringMap: "_" + t + "ToString",
		String2TypeMap: "_" + t + "FromString",
	}
}
