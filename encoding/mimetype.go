// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package encoding

import (
	"sort"
	"strings"

	"github.com/issue9/web/internal/accept"
)

// DefaultMimeType 默认的媒体类型，在不能正确获取输入和输出的媒体类型时，
// 会采用此值作为其默认值。
const DefaultMimeType = "application/octet-stream"

var (
	marshals   = make([]*marshaler, 0, 10)
	unmarshals = make([]*unmarshaler, 0, 10)
)

type (
	// MarshalFunc 将一个对象转换成 []byte 内容时，所采用的接口。
	MarshalFunc func(v interface{}) ([]byte, error)

	// UnmarshalFunc 将客户端内容转换成一个对象时，所采用的接口。
	UnmarshalFunc func([]byte, interface{}) error

	marshaler struct {
		f    MarshalFunc
		name string
	}

	unmarshaler struct {
		f    UnmarshalFunc
		name string
	}
)

func init() {
	// findMarshal 需要确保最少有一个元素在 marshals 中
	if err := AddMarshal(DefaultMimeType, TextMarshal); err != nil {
		panic(err)
	}

	// findUnmarshal 需要确保最少有一个元素在 unmarshals 中
	if err := AddUnmarshal(DefaultMimeType, TextUnmarshal); err != nil {
		panic(err)
	}
}

// AcceptMimeType 从 header 解析出当前请求所需要的解 mimetype 名称和对应的解码函数
//
// */* 表示匹配任意内容，一般会选择第一个元素作匹配；
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
func AcceptMimeType(header string) (string, MarshalFunc, error) {
	accepts, err := accept.Parse(header)
	if err != nil {
		return "", nil, err
	}

	for _, accept := range accepts {
		if m := findMarshal(accept.Value); m != nil {
			return m.name, m.f, nil
		}
	}

	return "", nil, ErrInvalidMimeType
}

// AddMarshal 添加编码函数
func AddMarshal(name string, m MarshalFunc) error {
	if strings.HasSuffix(name, "/*") {
		return ErrInvalidMimeType
	}

	for _, mt := range marshals {
		if mt.name == name {
			return ErrExists
		}
	}

	marshals = append(marshals, &marshaler{
		f:    m,
		name: name,
	})

	sort.SliceStable(marshals, func(i, j int) bool {
		return marshals[i].name < marshals[j].name
	})

	return nil
}

// AddUnmarshal 添加编码函数
func AddUnmarshal(name string, m UnmarshalFunc) error {
	if strings.HasSuffix(name, "/*") {
		return ErrInvalidMimeType
	}

	for _, mt := range unmarshals {
		if mt.name == name {
			return ErrExists
		}
	}

	unmarshals = append(unmarshals, &unmarshaler{
		f:    m,
		name: name,
	})

	sort.SliceStable(unmarshals, func(i, j int) bool {
		return unmarshals[i].name < unmarshals[j].name
	})

	return nil
}

func findMarshal(name string) *marshaler {
	switch {
	case name == "*/*":
		return marshals[0] // 由 init() 确保最少有一个元素
	case strings.HasSuffix(name, "/*"):
		prefix := name[:len(name)-3]
		for _, mt := range marshals {
			if strings.HasPrefix(mt.name, prefix) {
				return mt
			}
		}
	default:
		for _, mt := range marshals {
			if mt.name == name {
				return mt
			}
		}
	}
	return nil
}

func findUnmarshal(name string) *unmarshaler {
	switch {
	case name == "*/*":
		return unmarshals[0] // 由 init() 确保最少有一个元素
	case strings.HasSuffix(name, "/*"):
		prefix := name[:len(name)-3]
		for _, mt := range unmarshals {
			if strings.HasPrefix(mt.name, prefix) {
				return mt
			}
		}
	default:
		for _, mt := range unmarshals {
			if mt.name == name {
				return mt
			}
		}
	}
	return nil
}
