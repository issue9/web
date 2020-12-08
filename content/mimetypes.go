// SPDX-License-Identifier: MIT

package content

import (
	"errors"
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

// Nil 表示向客户端输出 nil 值
//
// 这是一个只有类型但是值为空的变量。在某些特殊情况下，
// 如果需要向客户端输出一个 nil 值的内容，可以使用此值。
var Nil *struct{}

var (
	// ErrNotFound 表示未找到指定名称的编解码函数
	//
	// 在 Mimetypes.Marshal 和 Mimetypes.Unmarshal 中会返回该错误。
	ErrNotFound = errors.New("未找到指定名称的编解码函数")

	// ErrExists 存在相同中名称的编解码函数
	//
	// 在 Mimetypes.AddMarshal 和 Mimetypes.AddUnmarshal 时如果已经存在相同名称，返回此错误。
	ErrExists = errors.New("已经存在相同名称的编解码函数")
)

// MarshalFunc 将一个对象转换成 []byte 内容时所采用的接口
type MarshalFunc func(v interface{}) ([]byte, error)

// UnmarshalFunc 将客户端内容转换成一个对象时所采用的接口
type UnmarshalFunc func([]byte, interface{}) error

type codec struct {
	name      string
	marshal   MarshalFunc
	unmarshal UnmarshalFunc
}

// Mimetypes 管理 mimetype 的增删改查
type Mimetypes struct {
	codecs []*codec
}

// NewMimetypes 返回 *Mimetypes 实例
func NewMimetypes() *Mimetypes {
	return &Mimetypes{
		codecs: make([]*codec, 0, 10),
	}
}

// ConentType 从 content-type 报头解析出需要用到的解码函数
func (mt *Mimetypes) ConentType(header string) (UnmarshalFunc, encoding.Encoding, error) {
	encName, charsetName, err := ParseContentType(header)
	if err != nil {
		return nil, nil, err
	}

	f, err := mt.Unmarshal(encName)
	if err != nil {
		return nil, nil, err
	}

	e, err := htmlindex.Get(charsetName)
	if err != nil {
		return nil, nil, err
	}

	return f, e, nil
}

// Unmarshal 查找指定名称的 UnmarshalFunc
func (mt *Mimetypes) Unmarshal(name string) (UnmarshalFunc, error) {
	for _, c := range mt.codecs {
		if c.name == name {
			return c.unmarshal, nil
		}
	}
	return nil, ErrNotFound
}

// Marshal 从 header 解析出当前请求所需要的解 mimetype 名称和对应的解码函数
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
func (mt *Mimetypes) Marshal(header string) (string, MarshalFunc, error) {
	if header == "" {
		if mm := mt.findMarshal("*/*"); mm != nil {
			return mm.name, mm.marshal, nil
		}
		return "", nil, ErrNotFound
	}

	accepts := qheader.Parse(header, "*/*")
	for _, accept := range accepts {
		if mm := mt.findMarshal(accept.Value); mm != nil {
			return mm.name, mm.marshal, nil
		}
	}

	return "", nil, ErrNotFound
}

// Add 添加编解码函数
//
// m 和 u 可以为 nil，表示仅作为一个占位符使用，具体处理要在 ServeHTTP 中另作处理。
func (mt *Mimetypes) Add(name string, m MarshalFunc, u UnmarshalFunc) error {
	if strings.IndexByte(name, '*') >= 0 {
		panic("name 不是一个有效的 mimetype 名称格式")
	}

	for _, c := range mt.codecs {
		if c.name == name {
			return ErrExists
		}
	}

	mt.codecs = append(mt.codecs, &codec{
		name:      name,
		marshal:   m,
		unmarshal: u,
	})

	sort.SliceStable(mt.codecs, func(i, j int) bool {
		if mt.codecs[i].name == DefaultMimetype {
			return true
		}

		if mt.codecs[j].name == DefaultMimetype {
			return false
		}

		return mt.codecs[i].name < mt.codecs[j].name
	})

	return nil
}

// Set 修改编解码函数
func (mt *Mimetypes) Set(name string, m MarshalFunc, u UnmarshalFunc) error {
	for _, c := range mt.codecs {
		if c.name == name {
			c.marshal = m
			c.unmarshal = u
			return nil
		}
	}

	return ErrNotFound
}

// Delete 删除指定名称的数据
func (mt *Mimetypes) Delete(name string) {
	size := sliceutil.Delete(mt.codecs, func(i int) bool {
		return mt.codecs[i].name == name
	})
	mt.codecs = mt.codecs[:size]
}

func (mt *Mimetypes) findMarshal(name string) *codec {
	switch {
	case len(mt.codecs) == 0:
		return nil
	case name == "" || name == "*/*":
		return mt.codecs[0] // 由 len(marshals) == 0 确保最少有一个元素
	case strings.HasSuffix(name, "/*"):
		prefix := name[:len(name)-3]
		for _, c := range mt.codecs {
			if strings.HasPrefix(c.name, prefix) {
				return c
			}
		}
	default:
		for _, c := range mt.codecs {
			if c.name == name {
				return c
			}
		}
	}
	return nil
}
