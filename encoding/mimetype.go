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

// 保存所有添加的 mimetype 类型，根表示 */* 的内容
var mimetypes = make([]*mimetype, 0, 10)

type (
	// MarshalFunc 将一个对象转换成 []byte 内容时，所采用的接口。
	MarshalFunc func(v interface{}) ([]byte, error)

	// UnmarshalFunc 将客户端内容转换成一个对象时，所采用的接口。
	UnmarshalFunc func([]byte, interface{}) error

	mimetype struct {
		marshal   MarshalFunc
		unmarshal UnmarshalFunc
		name      string
	}
)

func init() {
	err := AddMimeType(DefaultMimeType, TextMarshal, TextUnmarshal)
	if err != nil {
		panic(err)
	}
}

// AcceptMimeType 从 header 解析出当前请求所需要的解 mimetype 名称和对应的解码函数
//
// 不存在时，返回默认值，出错时，返回错误。
func AcceptMimeType(header string) (string, MarshalFunc, error) {
	accepts, err := accept.Parse(header)
	if err != nil {
		return "", nil, err
	}

	for _, accept := range accepts {
		if m := findMarshal(accept.Value); m != nil {
			return accept.Value, m, nil
		}
	}

	return "", nil, ErrUnsupportedMarshal
}

// AddMarshal 添加编码函数
//
// Deprecated: 改用 AddMimeType 代替
func AddMarshal(name string, m MarshalFunc) error {
	return AddMimeType(name, m, nil)
}

// AddUnmarshal 添加编码函数
//
// Deprecated: 改用 AddMimeType 代替
func AddUnmarshal(name string, m UnmarshalFunc) error {
	return AddMimeType(name, nil, m)
}

func (mt *mimetype) set(marshal MarshalFunc, unmarshal UnmarshalFunc) error {
	if mt.marshal != nil && marshal != nil {
		return ErrExists
	}
	mt.marshal = marshal

	if mt.unmarshal != nil && unmarshal != nil {
		return ErrExists
	}
	mt.unmarshal = unmarshal

	return nil
}

func findMarshal(name string) MarshalFunc {
	if name == "*/*" {
		return mimetypes[0].marshal
	}

	if strings.HasSuffix(name, "/*") {
		prefix := name[:len(name)-3]
		for _, mt := range mimetypes {
			if strings.HasPrefix(mt.name, prefix) && mt.marshal != nil {
				return mt.marshal
			}
		}
	}

	for _, mt := range mimetypes {
		if mt.name == name && mt.marshal != nil {
			return mt.marshal
		}
	}

	return nil
}

func findUnmarshal(name string) UnmarshalFunc {
	if name == "*/*" {
		return mimetypes[0].unmarshal
	}

	if strings.HasSuffix(name, "/*") {
		prefix := name[:len(name)-3]
		for _, mt := range mimetypes {
			if strings.HasPrefix(mt.name, prefix) && mt.unmarshal != nil {
				return mt.unmarshal
			}
		}
	}

	for _, mt := range mimetypes {
		if mt.name == name && mt.unmarshal != nil {
			return mt.unmarshal
		}
	}

	return nil
}

// AddMimeType 添加编码和解码方式
func AddMimeType(name string, marshal MarshalFunc, unmarshal UnmarshalFunc) error {
	for _, mt := range mimetypes {
		if mt.name == name &&
			((mt.marshal != nil && marshal != nil) || mt.unmarshal != nil && unmarshal != nil) {
			return ErrExists
		}
	}

	mimetypes = append(mimetypes, &mimetype{
		marshal:   marshal,
		unmarshal: unmarshal,
		name:      name,
	})

	sort.SliceStable(mimetypes, func(i, j int) bool {
		return mimetypes[i].name > mimetypes[j].name
	})

	return nil
}
