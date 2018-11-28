// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package encoding 提供了对编码的支持。
package encoding

import (
	"errors"
	"sort"
	"strings"

	"github.com/issue9/middleware/compress/accept"
)

// DefaultMimeType 默认的媒体类型，在不能获取输入和输出的媒体类型时，
// 会采用此值作为其默认值。
//
// 若编码函数中指定该类型的函数，则会使用该编码优先匹配 */* 等格式的请求。
const DefaultMimeType = "application/octet-stream"

var (
	marshals   = make([]*marshaler, 0, 10)
	unmarshals = make([]*unmarshaler, 0, 10)
)

var (
	// ErrExists 表示指定名称的项目已经存在。
	//
	// 在 AddCharset、Addmarshal 和 AddUnmarshal 中会返回此错误。
	ErrExists = errors.New("该名称的项目已经存在")

	// ErrInvalidMimeType 无效的 mimetype 值，一般为 content-type 或
	// Accept 等报头指定的 mimetype 值无效。
	ErrInvalidMimeType = errors.New("mimetype 无效")
)

// Nil 表示向客户端输出 nil 值。
//
// 这是一个只有类型但是值为空的变量。在某些特殊情况下，
// 如果需要向客户端输出一个 nil 值的内容，可以使用此值。
var Nil *struct{}

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

// Unmarshal 查找指定名称的 UnmarshalFunc
func Unmarshal(name string) (UnmarshalFunc, error) {
	var unmarshal *unmarshaler
	for _, mt := range unmarshals {
		if mt.name == name {
			unmarshal = mt
			break
		}
	}
	if unmarshal == nil {
		return nil, ErrInvalidMimeType
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
func Marshal(header string) (string, MarshalFunc, error) {
	if header == "" {
		if m := findMarshal("*/*"); m != nil {
			return m.name, m.f, nil
		}
		return "", nil, ErrInvalidMimeType
	}

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

// AddMarshals 添加多个编码函数
func AddMarshals(ms map[string]MarshalFunc) error {
	for k, v := range ms {
		if err := AddMarshal(k, v); err != nil {
			return err
		}
	}

	return nil
}

// AddMarshal 添加编码函数
func AddMarshal(name string, m MarshalFunc) error {
	if strings.HasSuffix(name, "/*") || name == "*" {
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
		if marshals[i].name == DefaultMimeType {
			return true
		}

		if marshals[j].name == DefaultMimeType {
			return false
		}

		return marshals[i].name < marshals[j].name
	})

	return nil
}

// AddUnmarshals 添加多个编码函数
func AddUnmarshals(ms map[string]UnmarshalFunc) error {
	for k, v := range ms {
		if err := AddUnmarshal(k, v); err != nil {
			return err
		}
	}

	return nil
}

// AddUnmarshal 添加编码函数
func AddUnmarshal(name string, m UnmarshalFunc) error {
	if strings.IndexByte(name, '*') >= 0 {
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
		if unmarshals[i].name == DefaultMimeType {
			return true
		}

		if unmarshals[j].name == DefaultMimeType {
			return false
		}

		return unmarshals[i].name < unmarshals[j].name
	})

	return nil
}

func findMarshal(name string) *marshaler {
	switch {
	case len(marshals) == 0:
		return nil
	case name == "" || name == "*/*":
		return marshals[0] // 由 len(marshals) == 0 确保最少有一个元素
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
