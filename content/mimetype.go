// SPDX-License-Identifier: MIT

package content

import (
	"fmt"
	"mime"
	"strings"

	"github.com/issue9/qheader"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"

	"github.com/issue9/web/serialization"
)

// DefaultMimetype 默认的媒体类型
//
// 在不能获取输入和输出的媒体类型时，会采用此值作为其默认值。
const DefaultMimetype = "application/octet-stream"

// Mimetypes 管理 mimetype 的序列化操作
func (c *Content) Mimetypes() *serialization.Serialization { return c.mimetypes }

// conentType 从 content-type 报头解析出需要用到的解码函数
func (c *Content) conentType(header string) (serialization.UnmarshalFunc, encoding.Encoding, error) {
	var (
		mt      = DefaultMimetype
		charset = DefaultCharset
	)

	if header != "" {
		mts, params, err := mime.ParseMediaType(header)
		if err != nil {
			return nil, nil, err
		}
		mt = mts
		if charset = params["charset"]; charset == "" {
			charset = DefaultCharset
		}
	}

	f, found := c.unmarshal(mt)
	if !found {
		return nil, nil, fmt.Errorf("未注册的解函数 %s", mt)
	}

	e, err := htmlindex.Get(charset)
	if err != nil {
		return nil, nil, err
	}

	return f, e, nil
}

// unmarshal 查找指定名称的 UnmarshalFunc
func (c *Content) unmarshal(name string) (serialization.UnmarshalFunc, bool) {
	name, _, u := c.Mimetypes().Search(name)
	return u, name != ""
}

// marshal 从 header 解析出当前请求所需要的 mimetype 名称和对应的解码函数
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
func (c *Content) marshal(header string) (string, serialization.MarshalFunc, bool) {
	if header == "" {
		if name, m := c.findMarshal("*/*"); name != "" {
			return name, m, true
		}
		return "", nil, false
	}

	accepts := qheader.Parse(header, "*/*")
	for _, accept := range accepts {
		if name, m := c.findMarshal(accept.Value); name != "" {
			return name, m, true
		}
	}

	return "", nil, false
}

func (c *Content) findMarshal(name string) (n string, m serialization.MarshalFunc) {
	switch {
	case c.Mimetypes().Len() == 0:
		return "", nil
	case name == "" || name == "*/*":
		n, m, _ = c.Mimetypes().SearchFunc(func(s string) bool { return true }) // 第一个元素
	case strings.HasSuffix(name, "/*"):
		prefix := name[:len(name)-3]
		n, m, _ = c.Mimetypes().SearchFunc(func(s string) bool { return strings.HasPrefix(s, prefix) })
	default:
		n, m, _ = c.Mimetypes().SearchFunc(func(s string) bool { return s == name })
	}
	return n, m
}
