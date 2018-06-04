// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package encoding

import (
	"testing"

	"github.com/issue9/assert"
)

func TestAcceptMimeType(t *testing.T) {
	a := assert.New(t)

	name, marshal, err := AcceptMimeType(DefaultMimeType)
	a.NotError(err).
		Equal(marshal, MarshalFunc(TextMarshal)).
		Equal(name, DefaultMimeType)

	// * 不指定，需要用户自行决定其表示方式
	name, marshal, err = AcceptMimeType("*/*")
	a.Error(err).
		Empty(name).
		Nil(marshal)

	name, marshal, err = AcceptMimeType("font/wotff")
	a.Error(err).
		Empty(name).
		Nil(marshal)
}

func TestMimeType(t *testing.T) {
	a := assert.New(t)

	m, um := MimeType(DefaultMimeType)
	a.NotNil(m).NotNil(um)

	// 并未指定 text/*
	m, um = MimeType("text")
	a.Nil(m).Nil(um)

	// 未指定 font
	m, um = MimeType("font")
	a.Nil(m).Nil(um)

	a.NotError(AddMimeType("text/*", TextMarshal, TextUnmarshal))
	m, um = MimeType("text")
	a.NotNil(m).NotNil(um)
	m, um = MimeType("text/html") // 不存在
	a.Nil(m).Nil(um)

	a.NotError(AddMimeType("text/html", TextMarshal, TextUnmarshal))
	m, um = MimeType("text/html")
	a.NotNil(m).NotNil(um)

	// 根元素未安装
	m, um = MimeType("*/*")
	a.Nil(m).Nil(um)

	a.NotError(AddMimeType("*/*", TextMarshal, TextUnmarshal))
	m, um = MimeType("*")
	a.NotNil(m).NotNil(um)

	a.Error(AddMimeType("", TextMarshal, TextUnmarshal))
}

func TestMimetype_set(t *testing.T) {
	a := assert.New(t)

	mt := &mimetype{}
	a.NotError(mt.set(TextMarshal, TextUnmarshal))

	a.ErrorType(mt.set(TextMarshal, nil), ErrExists)
	a.ErrorType(mt.set(nil, TextUnmarshal), ErrExists)
}

func TestParseMimeType(t *testing.T) {
	a := assert.New(t)

	name, subname := parseMimeType("*/*")
	a.Equal(name, "*").Equal(subname, "*")

	name, subname = parseMimeType("*/")
	a.Equal(name, "*").Equal(subname, "")

	name, subname = parseMimeType("*/text")
	a.Equal(name, "*").Equal(subname, "text")

	name, subname = parseMimeType("text/*")
	a.Equal(name, "text").Equal(subname, "*")

	name, subname = parseMimeType("text/")
	a.Equal(name, "text").Equal(subname, "")

	name, subname = parseMimeType("/")
	a.Equal(name, "/").Equal(subname, "")

	name, subname = parseMimeType("/*")
	a.Equal(name, "/*").Equal(subname, "")
}
