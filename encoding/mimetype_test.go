// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package encoding

import (
	"encoding/json"
	"testing"

	"github.com/issue9/assert"
)

func testMarshal(t *testing.T) {
	a := assert.New(t)

	a.Nil(Marshal("not exists"))
	a.NotNil(Marshal(DefaultMimeType))

	// 添加已存在的
	a.Equal(AddMarshal(DefaultMimeType, json.Marshal), ErrExists)

	a.NotError(AddMarshal("json", json.Marshal))
	a.NotNil(Marshal("json"))
}

func testUnmarshal(t *testing.T) {
	a := assert.New(t)

	a.Nil(Unmarshal("not exists"))
	a.NotNil(Unmarshal(DefaultMimeType))

	// 添加已存在的
	a.Equal(AddUnmarshal(DefaultMimeType, json.Unmarshal), ErrExists)

	a.NotError(AddUnmarshal("json", json.Unmarshal))
	a.NotNil(Unmarshal("json"))
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
