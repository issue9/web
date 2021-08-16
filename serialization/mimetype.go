// SPDX-License-Identifier: MIT

package serialization

import (
	"strings"

	"github.com/issue9/qheader"
)

// Mimetype 提供 mimetype 的序列化操作
type Mimetypes struct {
	*Serialization
}

// NewMimetypes 返回 Mimetypes 实例
func NewMimetypes(c int) *Mimetypes { return &Mimetypes{Serialization: New(c)} }

// UnmarshalFunc 查找指定名称的 UnmarshalFunc
func (ms *Mimetypes) UnmarshalFunc(name string) (UnmarshalFunc, bool) {
	name, _, u := ms.Search(name)
	return u, name != ""
}

// MarshalFunc 从 header 解析出当前请求所需要的 mimetype 名称和对应的解码函数
//
// */* 或是空值 表示匹配任意内容，一般会选择第一个元素作匹配；
// xx/* 表示匹配以 xx/ 开头的任意元素，一般会选择 xx/* 开头的第一个元素；
// xx/ 表示完全匹配以 xx/ 的内容
// 如果传递的内容如下：
//  application/json;q=0.9,*/*;q=1
// 则因为 */* 的 q 值比较高，而返回 */* 匹配的内容
//
// 在不完全匹配的情况下，返回值的名称依然是具体名称。
//  text/*;q=0.9
// 返回的名称可能是：
//  text/plain
func (ms *Mimetypes) MarshalFunc(header string) (string, MarshalFunc, bool) {
	if header == "" {
		if name, m := ms.findMarshal("*/*"); name != "" {
			return name, m, true
		}
		return "", nil, false
	}

	accepts := qheader.Parse(header, "*/*")
	for _, accept := range accepts {
		if name, m := ms.findMarshal(accept.Value); name != "" {
			return name, m, true
		}
	}

	return "", nil, false
}

func (ms *Mimetypes) findMarshal(name string) (n string, m MarshalFunc) {
	switch {
	case ms.Len() == 0:
		return "", nil
	case name == "" || name == "*/*":
		n, m, _ = ms.SearchFunc(func(s string) bool { return true }) // 第一个元素
	case strings.HasSuffix(name, "/*"):
		prefix := name[:len(name)-3]
		n, m, _ = ms.SearchFunc(func(s string) bool { return strings.HasPrefix(s, prefix) })
	default:
		n, m, _ = ms.SearchFunc(func(s string) bool { return s == name })
	}
	return n, m
}
