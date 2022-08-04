// SPDX-License-Identifier: MIT

package serialization

import (
	"strings"

	"github.com/issue9/localeutil"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"

	"github.com/issue9/web/internal/header"
	"github.com/issue9/web/serializer"
)

type Mimetypes struct {
	serializer.Serializer
}

func NewMimetypes(c int) *Mimetypes { return &Mimetypes{Serializer: New(c)} }

func (ms *Mimetypes) unmarshalFunc(name string) (serializer.UnmarshalFunc, bool) {
	name, _, u := ms.Search(name)
	return u, name != ""
}

// ContentType 从 content-type 报头中获取解码和字符集函数
//
// h 表示 content-type 报头的内容。如果字符集为 utf-8 或是未指定，返回的字符解码为 nil；
func (ms *Mimetypes) ContentType(h string) (serializer.UnmarshalFunc, encoding.Encoding, error) {
	mimetype, charset := header.ParseWithParam(h, "charset")

	f, found := ms.unmarshalFunc(mimetype)
	if !found {
		return nil, nil, localeutil.Error("not found serialization function for %s", mimetype)
	}

	if charset == "" || charset == "utf-8" {
		return f, nil, nil
	}
	e, err := htmlindex.Get(charset)
	if err != nil {
		return nil, nil, err
	}

	return f, e, nil
}

// MarshalFunc 从 h 解析出当前请求所需要的 mimetype 名称和对应的解码函数
//
// */* 或是空值 表示匹配任意内容，一般会选择第一个元素作匹配；
// xx/* 表示匹配以 xx/ 开头的任意元素，一般会选择 xx/* 开头的第一个元素；
// xx/ 表示完全匹配以 xx/ 的内容
// 如果传递的内容如下：
//
//	application/json;q=0.9,*/*;q=1
//
// 则因为 */* 的 q 值比较高，而返回 */* 匹配的内容
//
// 在不完全匹配的情况下，返回值的名称依然是具体名称。
//
//	text/*;q=0.9
//
// 返回的名称可能是：
//
//	text/plain
func (ms *Mimetypes) MarshalFunc(h string) (string, serializer.MarshalFunc, bool) {
	if h == "" {
		if name, m := ms.findMarshal("*/*"); name != "" {
			return name, m, true
		}
		return "", nil, false
	}

	items := header.ParseQHeader(h, "*/*")
	defer header.PutQHeader(&items)
	for _, item := range items {
		if name, m := ms.findMarshal(item.Value); name != "" {
			return name, m, true
		}
	}

	return "", nil, false
}

func (ms *Mimetypes) findMarshal(name string) (n string, m serializer.MarshalFunc) {
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
