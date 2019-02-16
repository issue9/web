// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mimetype

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/issue9/middleware/compress/accept"
)

// ErrNotFound 表示指定名称的 mimetype 解析函数未找到
var ErrNotFound = errors.New("未找到指定名称的 mimetype")

// Mimetypes 管理 mimetype 的处理函数
type Mimetypes struct {
	marshals   []*marshaler
	unmarshals []*unmarshaler
}

type marshaler struct {
	f    MarshalFunc
	name string
}

type unmarshaler struct {
	f    UnmarshalFunc
	name string
}

// New 声明新的 Mimetypes 实例
func New() *Mimetypes {
	return &Mimetypes{
		marshals:   make([]*marshaler, 0, 10),
		unmarshals: make([]*unmarshaler, 0, 10),
	}
}

func nameExists(name string) error {
	return fmt.Errorf("该名称 %s 已经存在", name)
}

// Unmarshal 查找指定名称的 UnmarshalFunc
func (m *Mimetypes) Unmarshal(name string) (UnmarshalFunc, error) {
	var unmarshal *unmarshaler
	for _, mt := range m.unmarshals {
		if mt.name == name {
			unmarshal = mt
			break
		}
	}
	if unmarshal == nil {
		return nil, ErrNotFound
	}

	return unmarshal.f, nil
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
func (m *Mimetypes) Marshal(header string) (string, MarshalFunc, error) {
	if header == "" {
		if mm := m.findMarshal("*/*"); mm != nil {
			return mm.name, mm.f, nil
		}
		return "", nil, ErrNotFound
	}

	accepts, err := accept.Parse(header)
	if err != nil {
		return "", nil, err
	}

	for _, accept := range accepts {
		if mm := m.findMarshal(accept.Value); mm != nil {
			return mm.name, mm.f, nil
		}
	}

	return "", nil, ErrNotFound
}

// AddMarshals 添加多个编码函数
func (m *Mimetypes) AddMarshals(ms map[string]MarshalFunc) error {
	for k, v := range ms {
		if err := m.AddMarshal(k, v); err != nil {
			return err
		}
	}

	return nil
}

// AddMarshal 添加编码函数
//
// mf 可以为 nil，表示仅作为一个占位符使用，具体处理要在 ServeHTTP
// 另作处理，比如下载，上传等内容。
func (m *Mimetypes) AddMarshal(name string, mf MarshalFunc) error {
	if strings.HasSuffix(name, "/*") || name == "*" {
		panic("name 不是一个有效的 mimetype 名称格式")
	}

	for _, mt := range m.marshals {
		if mt.name == name {
			return nameExists(name)
		}
	}

	m.marshals = append(m.marshals, &marshaler{
		f:    mf,
		name: name,
	})

	sort.SliceStable(m.marshals, func(i, j int) bool {
		if m.marshals[i].name == DefaultMimetype {
			return true
		}

		if m.marshals[j].name == DefaultMimetype {
			return false
		}

		return m.marshals[i].name < m.marshals[j].name
	})

	return nil
}

// AddUnmarshals 添加多个编码函数
func (m *Mimetypes) AddUnmarshals(ms map[string]UnmarshalFunc) error {
	for k, v := range ms {
		if err := m.AddUnmarshal(k, v); err != nil {
			return err
		}
	}

	return nil
}

// AddUnmarshal 添加编码函数
//
// mm 可以为 nil，表示仅作为一个占位符使用，具体处理要在 ServeHTTP
// 另作处理，比如下载，上传等内容。
func (m *Mimetypes) AddUnmarshal(name string, mm UnmarshalFunc) error {
	if strings.IndexByte(name, '*') >= 0 {
		panic("name 不是一个有效的 mimetype 名称格式")
	}

	for _, mt := range m.unmarshals {
		if mt.name == name {
			return nameExists(name)
		}
	}

	m.unmarshals = append(m.unmarshals, &unmarshaler{
		f:    mm,
		name: name,
	})

	sort.SliceStable(m.unmarshals, func(i, j int) bool {
		if m.unmarshals[i].name == DefaultMimetype {
			return true
		}

		if m.unmarshals[j].name == DefaultMimetype {
			return false
		}

		return m.unmarshals[i].name < m.unmarshals[j].name
	})

	return nil
}

func (m *Mimetypes) findMarshal(name string) *marshaler {
	switch {
	case len(m.marshals) == 0:
		return nil
	case name == "" || name == "*/*":
		return m.marshals[0] // 由 len(marshals) == 0 确保最少有一个元素
	case strings.HasSuffix(name, "/*"):
		prefix := name[:len(name)-3]
		for _, mt := range m.marshals {
			if strings.HasPrefix(mt.name, prefix) {
				return mt
			}
		}
	default:
		for _, mt := range m.marshals {
			if mt.name == name {
				return mt
			}
		}
	}
	return nil
}
