// SPDX-License-Identifier: MIT

package content

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/issue9/assert"
)

func TestContent_contentType(t *testing.T) {
	a := assert.New(t)

	mt := New(DefaultBuilder)
	a.NotNil(mt)

	f, e, err := mt.conentType(";;;")
	a.Error(err).Nil(f).Nil(e)

	// 不存在的 mimetype
	f, e, err = mt.conentType(buildContentType(DefaultMimetype, DefaultCharset))
	a.Error(err).Nil(f).Nil(e)

	mt.AddMimetype(nil, json.Unmarshal, DefaultMimetype)
	f, e, err = mt.conentType(buildContentType(DefaultMimetype, DefaultCharset))
	a.NotError(err).NotNil(f).NotNil(e)

	// 无效的字符集名称
	f, e, err = mt.conentType(buildContentType(DefaultMimetype, "invalid-charset"))
	a.Error(err).Nil(f).Nil(e)
}

func TestContent_unmarshal(t *testing.T) {
	a := assert.New(t)

	mt := New(DefaultBuilder)
	a.NotNil(mt)

	um, found := mt.unmarshal("")
	a.False(found).Nil(um)

	a.NotError(mt.AddMimetype(json.Marshal, json.Unmarshal, DefaultMimetype))

	um, found = mt.unmarshal(DefaultMimetype)
	a.True(found).NotNil(um)

	// 未指定 mimetype
	um, found = mt.unmarshal("")
	a.False(found).Nil(um)

	// mimetype 无法找到
	um, found = mt.unmarshal("not-exists")
	a.False(found).Nil(um)

	// 空的 unmarshal
	a.NotError(mt.AddMimetype(json.Marshal, nil, "empty"))
	um, found = mt.unmarshal("empty")
	a.True(found).Nil(um)
}

func TestContent_marshal(t *testing.T) {
	a := assert.New(t)
	mt := New(DefaultBuilder)

	name, marshal, found := mt.marshal(DefaultMimetype)
	a.False(found).
		Nil(marshal).
		Empty(name)

	name, marshal, found = mt.marshal("")
	a.False(found).
		Nil(marshal).
		Empty(name)

	a.NotError(mt.AddMimetype(xml.Marshal, xml.Unmarshal, DefaultMimetype))
	a.NotError(mt.AddMimetype(json.Marshal, json.Unmarshal, "text/plain"))
	a.NotError(mt.AddMimetype(nil, nil, "empty"))

	name, marshal, found = mt.marshal(DefaultMimetype)
	a.True(found).
		Equal(marshal, MarshalFunc(xml.Marshal)).
		Equal(name, DefaultMimetype)

	a.NotError(mt.SetMimetype(DefaultMimetype, json.Marshal, json.Unmarshal))
	name, marshal, found = mt.marshal(DefaultMimetype)
	a.True(found).
		Equal(marshal, MarshalFunc(json.Marshal)).
		Equal(name, DefaultMimetype)

	a.ErrorString(mt.SetMimetype("not-exists", nil, nil), "未找到指定名称")

	// */* 如果指定了 DefaultMimetype，则必定是该值
	name, marshal, found = mt.marshal("*/*")
	a.True(found).
		Equal(marshal, MarshalFunc(json.Marshal)).
		Equal(name, DefaultMimetype)

	// 同 */*
	name, marshal, found = mt.marshal("")
	a.True(found).
		Equal(marshal, MarshalFunc(json.Marshal)).
		Equal(name, DefaultMimetype)

	name, marshal, found = mt.marshal("*/*,text/plain")
	a.True(found).
		Equal(marshal, MarshalFunc(json.Marshal)).
		Equal(name, "text/plain")

	name, marshal, found = mt.marshal("font/wottf;q=x.9")
	a.False(found).
		Empty(name).
		Nil(marshal)

	name, marshal, found = mt.marshal("font/wottf")
	a.False(found).
		Empty(name).
		Nil(marshal)

	// 匹配 empty
	name, marshal, found = mt.marshal("empty")
	a.True(found).
		Equal(name, "empty").
		Nil(marshal)
}

func TestContent_Add_Delete(t *testing.T) {
	a := assert.New(t)
	mt := New(DefaultBuilder)
	a.NotNil(mt)

	// 不能添加同名的多次
	a.NotError(mt.AddMimetype(nil, nil, DefaultMimetype))
	a.ErrorString(mt.AddMimetype(nil, nil, DefaultMimetype), "已经存在相同名称")

	// 不能添加以 /* 结属的名称
	a.Panic(func() {
		a.NotError(mt.AddMimetype(nil, nil, "application/*"))
	})
	a.Panic(func() {
		a.NotError(mt.AddMimetype(nil, nil, "/*"))
	})

	// 排序是否正常
	a.NotError(mt.AddMimetype(nil, nil, "application/json"))
	a.Equal(mt.mimetypes[0].name, DefaultMimetype) // 默认始终在第一

	a.NotError(mt.AddMimetype(nil, nil, "text", "text/plain", "text/text"))
	a.NotError(mt.AddMimetype(nil, nil))                   // 缺少 name 参数，不会添加任何内容
	a.NotError(mt.AddMimetype(nil, nil, "application/aa")) // aa 排名靠前
	a.NotError(mt.AddMimetype(nil, nil, "application/bb"))

	// 检测排序
	a.Equal(mt.mimetypes[0].name, DefaultMimetype)
	a.Equal(mt.mimetypes[1].name, "application/aa")
	a.Equal(mt.mimetypes[2].name, "application/bb")
	a.Equal(mt.mimetypes[3].name, "application/json")
	a.Equal(mt.mimetypes[4].name, "text")
	a.Equal(mt.mimetypes[5].name, "text/plain")
	a.Equal(mt.mimetypes[6].name, "text/text")

	// 删除
	mt.DeleteMimetype("text")
	mt.DeleteMimetype(DefaultMimetype)
	mt.DeleteMimetype("not-exists")
	a.Equal(mt.mimetypes[0].name, "application/aa")
	a.Equal(mt.mimetypes[1].name, "application/bb")
	a.Equal(mt.mimetypes[2].name, "application/json")
	a.Equal(mt.mimetypes[3].name, "text/plain")
	a.Equal(mt.mimetypes[4].name, "text/text")
}

func TestContent_findMarshal(t *testing.T) {
	a := assert.New(t)
	mt := New(DefaultBuilder)

	a.NotError(mt.AddMimetype(nil, nil, "text", "text/plain", "text/text"))
	a.NotError(mt.AddMimetype(nil, nil, "application/aa")) // aa 排名靠前
	a.NotError(mt.AddMimetype(nil, nil, "application/bb"))

	mm := mt.findMarshal("text")
	a.Equal(mm.name, "text")

	mm = mt.findMarshal("text/*")
	a.Equal(mm.name, "text")

	mm = mt.findMarshal("application/*")
	a.Equal(mm.name, "application/aa")

	// 第一条数据
	mm = mt.findMarshal("*/*")
	a.Equal(mm.name, "application/aa")

	// 第一条数据
	mm = mt.findMarshal("")
	a.Equal(mm.name, "application/aa")

	// 有默认值，则始终在第一
	a.NotError(mt.AddMimetype(nil, nil, DefaultMimetype))
	mm = mt.findMarshal("*/*")
	a.Equal(mm.name, DefaultMimetype)

	// 不存在
	a.Nil(mt.findMarshal("xx/*"))
}
