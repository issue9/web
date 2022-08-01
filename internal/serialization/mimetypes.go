// SPDX-License-Identifier: MIT

package serialization

import (
	"mime"
	"strings"

	"github.com/issue9/localeutil"
	"github.com/issue9/qheader"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"

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
// header 表示 content-type 报头的内容；
// mimetype 表示默认解码名称，当 header 为空时采用此值；
// charset 表示默认字符集，当 header 中未指定时，采用此值；
func (ms *Mimetypes) ContentType(header, mimetype, charset string) (serializer.UnmarshalFunc, encoding.Encoding, error) {
	if header != "" {
		m, ps, err := mime.ParseMediaType(header)
		if err != nil {
			return nil, nil, err
		}
		mimetype = m

		if c := ps["charset"]; c != "" {
			charset = c
		}
	}

	f, found := ms.unmarshalFunc(mimetype)
	if !found {
		return nil, nil, localeutil.Error("not found serialization function for %s", mimetype)
	}

	e, err := htmlindex.Get(charset)
	if err != nil {
		return nil, nil, err
	}

	return f, e, nil
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
func (ms *Mimetypes) MarshalFunc(header string) (string, serializer.MarshalFunc, bool) {
	if header == "" {
		if name, m := ms.findMarshal("*/*"); name != "" {
			return name, m, true
		}
		return "", nil, false
	}

	if qh := qheader.Parse(header, "*/*"); qh != nil {
		for _, item := range qh.Items {
			if name, m := ms.findMarshal(item.Value); name != "" {
				return name, m, true
			}
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
