// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package mimetype

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/web/mimetype/gob"
)

func TestMimetypes_Unmarshal(t *testing.T) {
	a := assert.New(t)

	m := New()
	um, err := m.Unmarshal("")
	a.Error(err).
		Nil(um)

	a.NotError(m.AddUnmarshal(DefaultMimetype, gob.Unmarshal))
	a.NotError(m.AddMarshal(DefaultMimetype, gob.Marshal))

	// 未指定 mimetype
	um, err = m.Unmarshal("")
	a.Error(err).Nil(um)

	// mimetype 无法找到
	um, err = m.Unmarshal("not-exists")
	a.Error(err).Nil(um)
}

func TestMimetypes_Marshal(t *testing.T) {
	a := assert.New(t)
	m := New()

	name, marshal, err := m.Marshal(DefaultMimetype)
	a.Error(err).
		Nil(marshal).
		Empty(name)

	a.NotError(m.AddMarshal(DefaultMimetype, gob.Marshal))
	a.NotError(m.AddMarshal("text/plain", gob.Marshal))

	name, marshal, err = m.Marshal(DefaultMimetype)
	a.NotError(err).
		Equal(marshal, MarshalFunc(gob.Marshal)).
		Equal(name, DefaultMimetype)

	name, marshal, err = m.Marshal(DefaultMimetype)
	a.NotError(err).
		Equal(marshal, MarshalFunc(gob.Marshal)).
		Equal(name, DefaultMimetype)

	// */* 如果指定了 DefaultMimetype，则必定是该值
	name, marshal, err = m.Marshal("*/*")
	a.NotError(err).
		Equal(marshal, MarshalFunc(gob.Marshal)).
		Equal(name, DefaultMimetype)

	// 同 */*
	name, marshal, err = m.Marshal("")
	a.NotError(err).
		Equal(marshal, MarshalFunc(gob.Marshal)).
		Equal(name, DefaultMimetype)

	name, marshal, err = m.Marshal("*/*,text/plain")
	a.NotError(err).
		Equal(marshal, MarshalFunc(gob.Marshal)).
		Equal(name, "text/plain")

	name, marshal, err = m.Marshal("font/wottf;q=x.9")
	a.Error(err).
		Empty(name).
		Nil(marshal)

	name, marshal, err = m.Marshal("font/wottf")
	a.Error(err).
		Empty(name).
		Nil(marshal)
}

func TestMimetypes_AddMarshal(t *testing.T) {
	a := assert.New(t)
	m := New()

	// 不能添加同名的多次
	a.NotError(m.AddMarshal(DefaultMimetype, nil))
	a.Error(m.AddMarshal(DefaultMimetype, nil))

	// 不能添加以 /* 结属的名称
	a.Panic(func() {
		a.NotError(m.AddMarshal("application/*", nil))
	})
	a.Panic(func() {
		a.NotError(m.AddMarshal("/*", nil))
	})

	// 排序是否正常
	a.NotError(m.AddMarshal("application/json", nil))
	a.Equal(m.marshals[0].name, DefaultMimetype) // 默认始终在第一
}

func TestMimetypes_AddUnmarshal(t *testing.T) {
	a := assert.New(t)
	m := New()

	a.NotError(m.AddUnmarshal(DefaultMimetype, nil))
	a.Error(m.AddUnmarshal(DefaultMimetype, nil))

	// 不能添加包含 * 字符的名称
	a.Panic(func() {
		a.NotError(m.AddUnmarshal("application/*", nil))
	})
	a.Panic(func() {
		a.NotError(m.AddUnmarshal("*", nil))
	})

	// 排序是否正常
	a.NotError(m.AddUnmarshal("application/json", nil))
	a.Equal(m.unmarshals[0].name, DefaultMimetype) // 默认始终在第一
}

func TestMimetypes_findMarshal(t *testing.T) {
	a := assert.New(t)
	m := New()

	a.NotError(m.AddMarshals(map[string]MarshalFunc{
		"text":           nil,
		"text/plain":     nil,
		"text/text":      nil,
		"application/aa": nil, // aa 排名靠前
		"application/bb": nil, // aa 排名靠前
	}))

	mm := m.findMarshal("text")
	a.Equal(mm.name, "text")

	mm = m.findMarshal("text/*")
	a.Equal(mm.name, "text")

	mm = m.findMarshal("application/*")
	a.Equal(mm.name, "application/aa")

	// 第一条数据
	mm = m.findMarshal("*/*")
	a.Equal(mm.name, "application/aa")

	// 第一条数据
	mm = m.findMarshal("")
	a.Equal(mm.name, "application/aa")

	// 有默认值，则始终在第一
	a.NotError(m.AddMarshal(DefaultMimetype, nil))
	mm = m.findMarshal("*/*")
	a.Equal(mm.name, DefaultMimetype)

	// 不存在
	a.Nil(m.findMarshal("xx/*"))
}
