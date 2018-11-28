// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package encoding

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/web/encoding/gob"
)

func resetMarshals() {
	marshals = []*marshaler{}
}

func resetUnmarshals() {
	unmarshals = []*unmarshaler{}
}

func TestUnmarshal(t *testing.T) {
	a := assert.New(t)

	um, err := Unmarshal("")
	a.Error(err).
		Nil(um)

	a.NotError(AddUnmarshal(DefaultMimeType, gob.Unmarshal))
	a.NotError(AddMarshal(DefaultMimeType, gob.Marshal))

	// 未指定 mimetype
	um, err = Unmarshal("")
	a.Error(err).Nil(um)

	// mimetype 无法找到
	um, err = Unmarshal("not-exists")
	a.Error(err).Nil(um)
}

func TestAcceptMimeType(t *testing.T) {
	a := assert.New(t)
	resetMarshals()
	resetUnmarshals()

	name, marshal, err := AcceptMimeType(DefaultMimeType)
	a.Error(err).
		Nil(marshal).
		Empty(name)

	a.NotError(AddMarshal(DefaultMimeType, gob.Marshal))
	a.NotError(AddMarshal("text/plain", gob.Marshal))

	name, marshal, err = AcceptMimeType(DefaultMimeType)
	a.NotError(err).
		Equal(marshal, MarshalFunc(gob.Marshal)).
		Equal(name, DefaultMimeType)

	name, marshal, err = AcceptMimeType(DefaultMimeType)
	a.NotError(err).
		Equal(marshal, MarshalFunc(gob.Marshal)).
		Equal(name, DefaultMimeType)

	// */* 如果指定了 DefaultMimeType，则必定是该值
	name, marshal, err = AcceptMimeType("*/*")
	a.NotError(err).
		Equal(marshal, MarshalFunc(gob.Marshal)).
		Equal(name, DefaultMimeType)

	// 同 */*
	name, marshal, err = AcceptMimeType("")
	a.NotError(err).
		Equal(marshal, MarshalFunc(gob.Marshal)).
		Equal(name, DefaultMimeType)

	name, marshal, err = AcceptMimeType("*/*,text/plain")
	a.NotError(err).
		Equal(marshal, MarshalFunc(gob.Marshal)).
		Equal(name, "text/plain")

	name, marshal, err = AcceptMimeType("font/wotff;q=x.9")
	a.Error(err).
		Empty(name).
		Nil(marshal)

	name, marshal, err = AcceptMimeType("font/wotff")
	a.Error(err).
		Empty(name).
		Nil(marshal)
}

func TestAddMarshal(t *testing.T) {
	a := assert.New(t)
	resetMarshals()

	// 不能添加同名的多次
	a.NotError(AddMarshal(DefaultMimeType, nil))
	a.ErrorType(AddMarshal(DefaultMimeType, nil), ErrExists)

	// 不能添加以 /* 结属的名称
	a.ErrorType(AddMarshal("application/*", nil), ErrExists)
	a.ErrorType(AddMarshal("/*", nil), ErrExists)

	// 排序是否正常
	a.NotError(AddMarshal("application/json", nil))
	a.Equal(marshals[0].name, DefaultMimeType) // 默认始终在第一
}

func TestAddUnmarshal(t *testing.T) {
	a := assert.New(t)
	resetUnmarshals()

	a.NotError(AddUnmarshal(DefaultMimeType, nil))
	a.ErrorType(AddUnmarshal(DefaultMimeType, nil), ErrExists)

	// 不能添加包含 * 字符的名称
	a.ErrorType(AddUnmarshal("application/*", nil), ErrExists)
	a.ErrorType(AddUnmarshal("*", nil), ErrExists)

	// 排序是否正常
	a.NotError(AddUnmarshal("application/json", nil))
	a.Equal(unmarshals[0].name, DefaultMimeType) // 默认始终在第一
}

func TestFindMarshal(t *testing.T) {
	a := assert.New(t)
	resetMarshals()

	a.NotError(AddMarshals(map[string]MarshalFunc{
		"text":           nil,
		"text/plain":     nil,
		"text/text":      nil,
		"application/aa": nil, // aa 排名靠前
		"application/bb": nil, // aa 排名靠前
	}))

	m := findMarshal("text")
	a.Equal(m.name, "text")

	m = findMarshal("text/*")
	a.Equal(m.name, "text")

	m = findMarshal("application/*")
	a.Equal(m.name, "application/aa")

	// 第一条数据
	m = findMarshal("*/*")
	a.Equal(m.name, "application/aa")

	// 第一条数据
	m = findMarshal("")
	a.Equal(m.name, "application/aa")

	// 有默认值，则始终在第一
	a.NotError(AddMarshal(DefaultMimeType, nil))
	m = findMarshal("*/*")
	a.Equal(m.name, DefaultMimeType)

	// 不存在
	a.Nil(findMarshal("xx/*"))
}
