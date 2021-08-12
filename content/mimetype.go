// SPDX-License-Identifier: MIT

package content

import (
	"fmt"
	"mime"
	"sort"
	"strings"

	"github.com/issue9/qheader"
	"github.com/issue9/sliceutil"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/htmlindex"
)

// DefaultMimetype 默认的媒体类型
//
// 在不能获取输入和输出的媒体类型时， 会采用此值作为其默认值。
//
// 若编码函数中指定该类型的函数，则会使用该编码优先匹配 */* 等格式的请求。
const DefaultMimetype = "application/octet-stream"

type (
	// MarshalFunc 序列化函数原型
	MarshalFunc func(v interface{}) ([]byte, error)

	// UnmarshalFunc 反序列化函数原型
	UnmarshalFunc func([]byte, interface{}) error

	mimetype struct {
		name      string
		marshal   MarshalFunc
		unmarshal UnmarshalFunc
	}
)

// conentType 从 content-type 报头解析出需要用到的解码函数
func (c *Content) conentType(header string) (UnmarshalFunc, encoding.Encoding, error) {
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

// MimetypeFunc 以指定函数查找是否有符合要求的序列化函数
//
// match 为匹配函数，原型为 func(name string)bool，name 为序列化函数的名称，
// 用户可以判断此值是否符合要求，返回 true，表示匹配，会中断后续的匹配。
func (c *Content) MimetypeFunc(match func(string) bool) (MarshalFunc, UnmarshalFunc, bool) {
	if m := c.mimetypeFunc(match); m != nil {
		return m.marshal, m.unmarshal, true
	}
	return nil, nil, false
}

// Mimetype 查找指定名称的序列化方法
func (c *Content) Mimetype(name string) (MarshalFunc, UnmarshalFunc, bool) {
	return c.MimetypeFunc(func(n string) bool { return n == name })
}

// unmarshal 查找指定名称的 UnmarshalFunc
func (c *Content) unmarshal(name string) (UnmarshalFunc, bool) {
	_, u, found := c.Mimetype(name)
	return u, found
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
func (c *Content) marshal(header string) (string, MarshalFunc, bool) {
	if header == "" {
		if mm := c.findMarshal("*/*"); mm != nil {
			return mm.name, mm.marshal, true
		}
		return "", nil, false
	}

	accepts := qheader.Parse(header, "*/*")
	for _, accept := range accepts {
		if mm := c.findMarshal(accept.Value); mm != nil {
			return mm.name, mm.marshal, true
		}
	}

	return "", nil, false
}

// AddMimetype 添加序列化函数
//
// m 和 u 可以为 nil，表示仅作为一个占位符使用，具体处理要在 ServeHTTP 中另作处理；
//
// name 表示名称，一般为 mimetype 名称，比如 application/xml 等，用户也可以添加其它值，比如：
//  c.AddMimetype(json.Marshal, json.Unmarshal, "application/json", ".json")
// 后期用户可以根据文件后缀名从 c.Mimetype 直接查找相应的序列化函数。
func (c *Content) AddMimetype(m MarshalFunc, u UnmarshalFunc, name ...string) error {
	for _, n := range name {
		if err := c.addMimetype(n, m, u); err != nil {
			return err
		}
	}
	return nil
}

func (c *Content) addMimetype(name string, m MarshalFunc, u UnmarshalFunc) error {
	if strings.IndexByte(name, '*') >= 0 {
		panic("name 不是一个有效的 mimetype 名称格式")
	}

	for _, mt := range c.mimetypes {
		if mt.name == name {
			return fmt.Errorf("已经存在相同名称 %s 的序列化函数", name)
		}
	}

	c.mimetypes = append(c.mimetypes, &mimetype{
		name:      name,
		marshal:   m,
		unmarshal: u,
	})

	sort.SliceStable(c.mimetypes, func(i, j int) bool {
		if c.mimetypes[i].name == DefaultMimetype {
			return true
		}

		if c.mimetypes[j].name == DefaultMimetype {
			return false
		}

		return c.mimetypes[i].name < c.mimetypes[j].name
	})

	return nil
}

// SetMimetype 修改编解码函数
func (c *Content) SetMimetype(name string, m MarshalFunc, u UnmarshalFunc) error {
	for _, mt := range c.mimetypes {
		if mt.name == name {
			mt.marshal = m
			mt.unmarshal = u
			return nil
		}
	}

	return fmt.Errorf("未找到指定名称 %s 的编解码函数", name)
}

// DeleteMimetype 删除指定名称的数据
func (c *Content) DeleteMimetype(name string) {
	size := sliceutil.Delete(c.mimetypes, func(i int) bool {
		return c.mimetypes[i].name == name
	})
	c.mimetypes = c.mimetypes[:size]
}

func (c *Content) findMarshal(name string) *mimetype {
	switch {
	case len(c.mimetypes) == 0:
		return nil
	case name == "" || name == "*/*":
		return c.mimetypes[0] // 由 len(marshals) == 0 确保最少有一个元素
	case strings.HasSuffix(name, "/*"):
		prefix := name[:len(name)-3]
		return c.mimetypeFunc(func(s string) bool { return strings.HasPrefix(s, prefix) })
	default:
		return c.mimetypeFunc(func(s string) bool { return s == name })
	}
}

func (c *Content) mimetypeFunc(match func(string) bool) *mimetype {
	for _, mt := range c.mimetypes {
		if match(mt.name) {
			return mt
		}
	}
	return nil
}
