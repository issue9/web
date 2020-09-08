// SPDX-License-Identifier: MIT

package context

import (
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/web/context/mimetype"
	"github.com/issue9/web/context/mimetype/gob"
)

func TestMimetypes_Unmarshal(t *testing.T) {
	a := assert.New(t)

	b := newEmptyBuilder(a)
	um, err := b.unmarshal("")
	a.Error(err).
		Nil(um)

	a.NotError(b.AddUnmarshal(mimetype.DefaultMimetype, gob.Unmarshal))
	a.NotError(b.AddMarshal(mimetype.DefaultMimetype, gob.Marshal))

	// 未指定 mimetype
	um, err = b.unmarshal("")
	a.Error(err).Nil(um)

	// mimetype 无法找到
	um, err = b.unmarshal("not-exists")
	a.Error(err).Nil(um)
}

func TestMimetypes_Marshal(t *testing.T) {
	a := assert.New(t)
	b := newEmptyBuilder(a)

	name, marshal, err := b.marshal(mimetype.DefaultMimetype)
	a.Error(err).
		Nil(marshal).
		Empty(name)

	name, marshal, err = b.marshal("")
	a.ErrorString(err, "请求中未指定 accept 报头，且服务端也未指定匹配 */* 的解码函数").
		Nil(marshal).
		Empty(name)

	a.NotError(b.AddMarshal(mimetype.DefaultMimetype, gob.Marshal))
	a.NotError(b.AddMarshal("text/plain", gob.Marshal))

	name, marshal, err = b.marshal(mimetype.DefaultMimetype)
	a.NotError(err).
		Equal(marshal, mimetype.MarshalFunc(gob.Marshal)).
		Equal(name, mimetype.DefaultMimetype)

	name, marshal, err = b.marshal(mimetype.DefaultMimetype)
	a.NotError(err).
		Equal(marshal, mimetype.MarshalFunc(gob.Marshal)).
		Equal(name, mimetype.DefaultMimetype)

	// */* 如果指定了 DefaultMimetype，则必定是该值
	name, marshal, err = b.marshal("*/*")
	a.NotError(err).
		Equal(marshal, mimetype.MarshalFunc(gob.Marshal)).
		Equal(name, mimetype.DefaultMimetype)

	// 同 */*
	name, marshal, err = b.marshal("")
	a.NotError(err).
		Equal(marshal, mimetype.MarshalFunc(gob.Marshal)).
		Equal(name, mimetype.DefaultMimetype)

	name, marshal, err = b.marshal("*/*,text/plain")
	a.NotError(err).
		Equal(marshal, mimetype.MarshalFunc(gob.Marshal)).
		Equal(name, "text/plain")

	name, marshal, err = b.marshal("font/wottf;q=x.9")
	a.Error(err).
		Empty(name).
		Nil(marshal)

	name, marshal, err = b.marshal("font/wottf")
	a.Error(err).
		Empty(name).
		Nil(marshal)
}

func TestMimetypes_AddMarshal(t *testing.T) {
	a := assert.New(t)
	b := newEmptyBuilder(a)

	// 不能添加同名的多次
	a.NotError(b.AddMarshal(mimetype.DefaultMimetype, nil))
	a.Error(b.AddMarshal(mimetype.DefaultMimetype, nil))

	// 不能添加以 /* 结属的名称
	a.Panic(func() {
		a.NotError(b.AddMarshal("application/*", nil))
	})
	a.Panic(func() {
		a.NotError(b.AddMarshal("/*", nil))
	})

	// 排序是否正常
	a.NotError(b.AddMarshal("application/json", nil))
	a.Equal(b.marshals[0].name, mimetype.DefaultMimetype) // 默认始终在第一
}

func TestMimetypes_AddUnmarshal(t *testing.T) {
	a := assert.New(t)
	builder := newEmptyBuilder(a)
	a.NotNil(builder)

	a.NotError(builder.AddUnmarshal(mimetype.DefaultMimetype, nil))
	a.Error(builder.AddUnmarshal(mimetype.DefaultMimetype, nil))

	// 不能添加包含 * 字符的名称
	a.Panic(func() {
		a.NotError(builder.AddUnmarshal("application/*", nil))
	})
	a.Panic(func() {
		a.NotError(builder.AddUnmarshal("*", nil))
	})

	// 排序是否正常
	a.NotError(builder.AddUnmarshal("application/json", nil))
	a.Equal(builder.unmarshals[0].name, mimetype.DefaultMimetype) // 默认始终在第一
}

func TestMimetypes_AddUnmarshals(t *testing.T) {
	a := assert.New(t)
	b := newEmptyBuilder(a)
	a.NotNil(b)

	err := b.AddUnmarshals(map[string]mimetype.UnmarshalFunc{
		mimetype.DefaultMimetype: nil,
		"text":                   nil,
		"application/json":       nil,
		"application/xml":        nil,
	})
	a.NotError(err)

	a.Equal(b.unmarshals[0].name, mimetype.DefaultMimetype)
	a.Equal(b.unmarshals[1].name, "application/json")
	a.Equal(b.unmarshals[2].name, "application/xml")
	a.Equal(b.unmarshals[3].name, "text")

	_, err = b.unmarshal("*/*")
	a.ErrorString(err, "未找到 */* 类型的解码函数")

	_, err = b.unmarshal("text")
	a.NotError(err)
}

func TestMimetypes_findMarshal(t *testing.T) {
	a := assert.New(t)
	b := newEmptyBuilder(a)

	a.NotError(b.AddMarshals(map[string]mimetype.MarshalFunc{
		"text":           nil,
		"text/plain":     nil,
		"text/text":      nil,
		"application/aa": nil, // aa 排名靠前
		"application/bb": nil, // aa 排名靠前
	}))

	// 检测排序
	a.Equal(b.marshals[0].name, "application/aa")
	a.Equal(b.marshals[1].name, "application/bb")
	a.Equal(b.marshals[2].name, "text")
	a.Equal(b.marshals[3].name, "text/plain")
	a.Equal(b.marshals[4].name, "text/text")

	mm := b.findMarshal("text")
	a.Equal(mm.name, "text")

	mm = b.findMarshal("text/*")
	a.Equal(mm.name, "text")

	mm = b.findMarshal("application/*")
	a.Equal(mm.name, "application/aa")

	// 第一条数据
	mm = b.findMarshal("*/*")
	a.Equal(mm.name, "application/aa")

	// 第一条数据
	mm = b.findMarshal("")
	a.Equal(mm.name, "application/aa")

	// 有默认值，则始终在第一
	a.NotError(b.AddMarshal(mimetype.DefaultMimetype, nil))
	mm = b.findMarshal("*/*")
	a.Equal(mm.name, mimetype.DefaultMimetype)

	// 不存在
	a.Nil(b.findMarshal("xx/*"))
}
