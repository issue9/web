// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package encoding

import (
	"errors"
	"strings"
)

// DefaultMimeType 默认的媒体类型，在不能正确获取输入和输出的媒体类型时，
// 会采用此值作为其默认值。
const DefaultMimeType = "application/octet-stream"

// 保存所有添加的 mimetype 类型，根表示 */* 的内容
var mimetypes = &mimetype{
	subtypes: make(map[string]*mimetype, 7),
}

type (
	// MarshalFunc 将一个对象转换成 []byte 内容时，所采用的接口。
	MarshalFunc func(v interface{}) ([]byte, error)

	// UnmarshalFunc 将客户端内容转换成一个对象时，所采用的接口。
	UnmarshalFunc func([]byte, interface{}) error

	mimetype struct {
		marshal   MarshalFunc
		unmarshal UnmarshalFunc
		subtypes  map[string]*mimetype
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
	accepts, err := ParseAccept(header)
	if err != nil {
		return "", nil, err
	}

	for _, accept := range accepts {
		if m, _ := MimeType(accept.Value); m != nil {
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

// MimeType 获取指定名称的编解码函数
func MimeType(name string) (MarshalFunc, UnmarshalFunc) {
	name, subname := parseMimeType(name)

	if name == "*" {
		return mimetypes.marshal, mimetypes.unmarshal
	}

	item, found := mimetypes.subtypes[name]
	if !found {
		return nil, nil
	}

	if subname == "" || subname == "*" {
		return item.marshal, item.unmarshal
	}

	sub, found := item.subtypes[subname]
	if !found {
		return nil, nil
	}
	return sub.marshal, sub.unmarshal
}

// AddMimeType 添加编码和解码方式
func AddMimeType(name string, marshal MarshalFunc, unmarshal UnmarshalFunc) error {
	if name == "" {
		return errors.New("参数 name 不能为空")
	}
	name, subname := parseMimeType(name)

	if name[0] == '*' {
		return mimetypes.set(marshal, unmarshal)
	}

	item, found := mimetypes.subtypes[name]
	if !found {
		item = &mimetype{
			subtypes: make(map[string]*mimetype, 10),
		}
		mimetypes.subtypes[name] = item
	}

	// 没有子名称
	if subname == "" || subname == "*" {
		return item.set(marshal, unmarshal)
	}

	sub, found := item.subtypes[subname]
	if !found {
		sub = &mimetype{} // 只有二级，所以子项不用再申请 subytpes 空间
		item.subtypes[subname] = sub
	}
	return sub.set(marshal, unmarshal)
}

func parseMimeType(mimename string) (name, subname string) {
	index := strings.IndexByte(mimename, '/')
	if index > 0 {
		return mimename[:index], mimename[index+1:]
	}
	return mimename, ""
}
