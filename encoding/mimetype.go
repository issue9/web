// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package encoding

import (
	"errors"
	"strings"
)

// 保存所有添加的 mimetype 类型，根表示 */* 的内容
var mimetypes = &mimetype{
	subtypes: make(map[string]*mimetype, 7),
}

type mimetype struct {
	marshal   MarshalFunc
	unmarshal UnmarshalFunc
	subtypes  map[string]*mimetype
}

func init() {
	err := AddMimetype(DefaultMimeType, TextMarshal, TextUnmarshal)
	if err != nil {
		panic(err)
	}
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

// Mimetype 获取指定名称的编解码函数
func Mimetype(name string) (MarshalFunc, UnmarshalFunc) {
	name, subname := parseMimetype(name)

	if name == "*" {
		return mimetypes.marshal, mimetypes.unmarshal
	}

	item, found := mimetypes.subtypes[name]
	if !found {
		return nil, nil
	}

	if subname == "" {
		return item.marshal, item.unmarshal
	}

	sub, found := item.subtypes[subname]
	if !found {
		return nil, nil
	}
	return sub.marshal, sub.unmarshal
}

// AddMimetype 添加编码和解码方式
func AddMimetype(name string, marshal MarshalFunc, unmarshal UnmarshalFunc) error {
	if name == "" {
		return errors.New("参数 name 不能为空")
	}
	name, subname := parseMimetype(name)

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
	if subname == "" {
		return item.set(marshal, unmarshal)
	}

	sub, found := item.subtypes[subname]
	if !found {
		sub = &mimetype{} // 只有二级，所以子项不用再申请 subytpes 空间
		item.subtypes[subname] = sub
	}
	return sub.set(marshal, unmarshal)
}

func parseMimetype(mimename string) (name, subname string) {
	index := strings.IndexByte(mimename, '/')
	if index > 0 {
		return mimename[:index], mimename[index+1:]
	}
	return mimename, ""
}
