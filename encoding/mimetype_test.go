// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package encoding

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/web/encoding/text"
)

func resetMarshals() {
	marshals = []*marshaler{
		&marshaler{name: DefaultMimeType, f: text.Marshal},
	}
}

func resetUnmarshals() {
	unmarshals = []*unmarshaler{
		&unmarshaler{name: DefaultMimeType, f: text.Unmarshal},
	}
}

func TestAcceptMimeType(t *testing.T) {
	a := assert.New(t)

	name, marshal, err := AcceptMimeType(DefaultMimeType)
	a.NotError(err).
		Equal(marshal, MarshalFunc(text.Marshal)).
		Equal(name, DefaultMimeType)

	// 匹配任意内容，一般选取第一个
	name, marshal, err = AcceptMimeType("*/*")
	a.NotError(err).
		Equal(marshal, MarshalFunc(text.Marshal)).
		Equal(name, DefaultMimeType)

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

	a.ErrorType(AddMarshal(DefaultMimeType, nil), ErrExists)

	// 不能添加以 /* 结属的名称
	a.ErrorType(AddMarshal("application/*", nil), ErrExists)
	a.ErrorType(AddMarshal("/*", nil), ErrExists)

	// 排序是否正常
	a.NotError(AddMarshal("application/json", nil))
	a.Equal(marshals[1].name, DefaultMimeType)
}

func TestAddUnmarshal(t *testing.T) {
	a := assert.New(t)

	a.ErrorType(AddUnmarshal(DefaultMimeType, nil), ErrExists)

	// 不能添加以 /* 结属的名称
	a.ErrorType(AddUnmarshal("application/*", nil), ErrExists)
	a.ErrorType(AddUnmarshal("/*", nil), ErrExists)

	// 排序是否正常
	a.NotError(AddUnmarshal("application/json", nil))
	a.Equal(unmarshals[1].name, DefaultMimeType)
}

func TestFindMarshal(t *testing.T) {
	a := assert.New(t)
	resetMarshals()

	a.NotError(AddMarshal("text", nil))
	a.NotError(AddMarshal("text/plain", nil))
	a.NotError(AddMarshal("text/text", nil))
	a.NotError(AddMarshal("application/json", nil))

	m := findMarshal("text")
	a.Equal(m.name, "text")

	m = findMarshal("text/*")
	a.Equal(m.name, "text")

	m = findMarshal("application/*")
	a.Equal(m.name, "application/json")

	m = findMarshal("*/*")
	a.Equal(m.name, "application/json")

	// 不存在
	a.Nil(findMarshal("xx/*"))
}

func TestFindUnmarshal(t *testing.T) {
	a := assert.New(t)
	resetUnmarshals()

	a.NotError(AddUnmarshal("text", nil))
	a.NotError(AddUnmarshal("text/plain", nil))
	a.NotError(AddUnmarshal("text/text", nil))
	a.NotError(AddUnmarshal("application/json", nil))

	m := findUnmarshal("text")
	a.Equal(m.name, "text")

	m = findUnmarshal("text/*")
	a.Equal(m.name, "text")

	m = findUnmarshal("application/*")
	a.Equal(m.name, "application/json")

	m = findUnmarshal("*/*")
	a.Equal(m.name, "application/json")

	// 不存在
	a.Nil(findUnmarshal("xx/*"))
}
