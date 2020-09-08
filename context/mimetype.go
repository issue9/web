// SPDX-License-Identifier: MIT

package context

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/issue9/qheader"

	"github.com/issue9/web/context/mimetype"
)

type marshaler struct {
	f    mimetype.MarshalFunc
	name string
}

type unmarshaler struct {
	f    mimetype.UnmarshalFunc
	name string
}

func mimetypeExists(name string) error {
	return fmt.Errorf("该名称 %s 的 mimetype 已经存在", name)
}

// unmarshal 查找指定名称的 UnmarshalFunc
func (b *Builder) unmarshal(name string) (mimetype.UnmarshalFunc, error) {
	var unmarshal *unmarshaler
	for _, mt := range b.unmarshals {
		if mt.name == name {
			unmarshal = mt
			break
		}
	}
	if unmarshal == nil {
		return nil, fmt.Errorf("未找到 %s 类型的解码函数", name)
	}

	return unmarshal.f, nil
}

// marshal 从 header 解析出当前请求所需要的解 mimetype 名称和对应的解码函数
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
func (b *Builder) marshal(header string) (string, mimetype.MarshalFunc, error) {
	if header == "" {
		if mm := b.findMarshal("*/*"); mm != nil {
			return mm.name, mm.f, nil
		}
		return "", nil, errors.New("请求中未指定 accept 报头，且服务端也未指定匹配 */* 的解码函数")
	}

	accepts, err := qheader.Parse(header, "*/*")
	if err != nil {
		return "", nil, err
	}

	for _, accept := range accepts {
		if mm := b.findMarshal(accept.Value); mm != nil {
			return mm.name, mm.f, nil
		}
	}

	return "", nil, errors.New("未找到符合客户端要求的解码函数")
}

// AddMarshals 添加多个编码函数
func (b *Builder) AddMarshals(ms map[string]mimetype.MarshalFunc) error {
	for k, v := range ms {
		if err := b.AddMarshal(k, v); err != nil {
			return err
		}
	}

	return nil
}

// AddMarshal 添加编码函数
//
// mf 可以为 nil，表示仅作为一个占位符使用，具体处理要在 ServeHTTP
// 另作处理，比如下载，上传等内容。
func (b *Builder) AddMarshal(name string, mf mimetype.MarshalFunc) error {
	if strings.HasSuffix(name, "/*") || name == "*" {
		panic("name 不是一个有效的 mimetype 名称格式")
	}

	for _, mt := range b.marshals {
		if mt.name == name {
			return mimetypeExists(name)
		}
	}

	b.marshals = append(b.marshals, &marshaler{
		f:    mf,
		name: name,
	})

	sort.SliceStable(b.marshals, func(i, j int) bool {
		if b.marshals[i].name == mimetype.DefaultMimetype {
			return true
		}

		if b.marshals[j].name == mimetype.DefaultMimetype {
			return false
		}

		return b.marshals[i].name < b.marshals[j].name
	})

	return nil
}

// AddUnmarshals 添加多个编码函数
func (b *Builder) AddUnmarshals(ms map[string]mimetype.UnmarshalFunc) error {
	for k, v := range ms {
		if err := b.AddUnmarshal(k, v); err != nil {
			return err
		}
	}

	return nil
}

// AddUnmarshal 添加编码函数
//
// mm 可以为 nil，表示仅作为一个占位符使用，具体处理要在 ServeHTTP
// 另作处理，比如下载，上传等内容。
func (b *Builder) AddUnmarshal(name string, mm mimetype.UnmarshalFunc) error {
	if strings.IndexByte(name, '*') >= 0 {
		panic("name 不是一个有效的 mimetype 名称格式")
	}

	for _, mt := range b.unmarshals {
		if mt.name == name {
			return mimetypeExists(name)
		}
	}

	b.unmarshals = append(b.unmarshals, &unmarshaler{
		f:    mm,
		name: name,
	})

	sort.SliceStable(b.unmarshals, func(i, j int) bool {
		if b.unmarshals[i].name == mimetype.DefaultMimetype {
			return true
		}

		if b.unmarshals[j].name == mimetype.DefaultMimetype {
			return false
		}

		return b.unmarshals[i].name < b.unmarshals[j].name
	})

	return nil
}

func (b *Builder) findMarshal(name string) *marshaler {
	switch {
	case len(b.marshals) == 0:
		return nil
	case name == "" || name == "*/*":
		return b.marshals[0] // 由 len(marshals) == 0 确保最少有一个元素
	case strings.HasSuffix(name, "/*"):
		prefix := name[:len(name)-3]
		for _, mt := range b.marshals {
			if strings.HasPrefix(mt.name, prefix) {
				return mt
			}
		}
	default:
		for _, mt := range b.marshals {
			if mt.name == name {
				return mt
			}
		}
	}
	return nil
}
