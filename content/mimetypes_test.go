// SPDX-License-Identifier: MIT

package content

import (
	"encoding/json"
	"testing"

	"github.com/issue9/assert"
)

func TestMimetypes_ContentType(t *testing.T) {
	a := assert.New(t)

	mt := NewMimetypes()
	a.NotNil(mt)

	f, e, err := mt.ConentType(";;;")
	a.Error(err).Nil(f).Nil(e)

	// 不存在的 mimetype
	f, e, err = mt.ConentType(BuildContentType(DefaultMimetype, DefaultCharset))
	a.Error(err).Nil(f).Nil(e)

	mt.Add(DefaultMimetype, json.Marshal, json.Unmarshal)
	f, e, err = mt.ConentType(BuildContentType(DefaultMimetype, DefaultCharset))
	a.NotError(err).NotNil(f).NotNil(e)

	// 无效的字符集名称
	f, e, err = mt.ConentType(BuildContentType(DefaultMimetype, "invalid-charset"))
	a.Error(err).Nil(f).Nil(e)
}

func TestMimetypes_Unmarshal(t *testing.T) {
	a := assert.New(t)

	mt := NewMimetypes()
	a.NotNil(mt)

	um, err := mt.Unmarshal("")
	a.Error(err).
		Nil(um)

	a.NotError(mt.Add(DefaultMimetype, json.Marshal, json.Unmarshal))

	um, err = mt.Unmarshal(DefaultMimetype)
	a.NotError(err).NotNil(um)

	// 未指定 mimetype
	um, err = mt.Unmarshal("")
	a.Error(err).Nil(um)

	// mimetype 无法找到
	um, err = mt.Unmarshal("not-exists")
	a.ErrorIs(err, ErrNotFound).Nil(um)
}

func TestMimetypes_Marshal(t *testing.T) {
	a := assert.New(t)
	mt := NewMimetypes()

	name, marshal, err := mt.Marshal(DefaultMimetype)
	a.Error(err).
		Nil(marshal).
		Empty(name)

	name, marshal, err = mt.Marshal("")
	a.ErrorIs(err, ErrNotFound).
		Nil(marshal).
		Empty(name)

	a.NotError(mt.Add(DefaultMimetype, json.Marshal, json.Unmarshal))
	a.NotError(mt.Add("text/plain", json.Marshal, json.Unmarshal))

	name, marshal, err = mt.Marshal(DefaultMimetype)
	a.NotError(err).
		Equal(marshal, MarshalFunc(json.Marshal)).
		Equal(name, DefaultMimetype)

	name, marshal, err = mt.Marshal(DefaultMimetype)
	a.NotError(err).
		Equal(marshal, MarshalFunc(json.Marshal)).
		Equal(name, DefaultMimetype)

	// */* 如果指定了 DefaultMimetype，则必定是该值
	name, marshal, err = mt.Marshal("*/*")
	a.NotError(err).
		Equal(marshal, MarshalFunc(json.Marshal)).
		Equal(name, DefaultMimetype)

	// 同 */*
	name, marshal, err = mt.Marshal("")
	a.NotError(err).
		Equal(marshal, MarshalFunc(json.Marshal)).
		Equal(name, DefaultMimetype)

	name, marshal, err = mt.Marshal("*/*,text/plain")
	a.NotError(err).
		Equal(marshal, MarshalFunc(json.Marshal)).
		Equal(name, "text/plain")

	name, marshal, err = mt.Marshal("font/wottf;q=x.9")
	a.Error(err).
		Empty(name).
		Nil(marshal)

	name, marshal, err = mt.Marshal("font/wottf")
	a.Error(err).
		Empty(name).
		Nil(marshal)
}

func TestMimetypes_Add(t *testing.T) {
	a := assert.New(t)
	mt := NewMimetypes()
	a.NotNil(mt)

	// 不能添加同名的多次
	a.NotError(mt.Add(DefaultMimetype, nil, nil))
	a.ErrorIs(mt.Add(DefaultMimetype, nil, nil), ErrExists)

	// 不能添加以 /* 结属的名称
	a.Panic(func() {
		a.NotError(mt.Add("application/*", nil, nil))
	})
	a.Panic(func() {
		a.NotError(mt.Add("/*", nil, nil))
	})

	// 排序是否正常
	a.NotError(mt.Add("application/json", nil, nil))
	a.Equal(mt.codecs[0].name, DefaultMimetype) // 默认始终在第一

	a.NotError(mt.Add("text", nil, nil))
	a.NotError(mt.Add("text/plain", nil, nil))
	a.NotError(mt.Add("text/text", nil, nil))
	a.NotError(mt.Add("application/aa", nil, nil)) // aa 排名靠前
	a.NotError(mt.Add("application/bb", nil, nil))

	// 检测排序
	a.Equal(mt.codecs[0].name, DefaultMimetype)
	a.Equal(mt.codecs[1].name, "application/aa")
	a.Equal(mt.codecs[2].name, "application/bb")
	a.Equal(mt.codecs[3].name, "application/json")
	a.Equal(mt.codecs[4].name, "text")
	a.Equal(mt.codecs[5].name, "text/plain")
	a.Equal(mt.codecs[6].name, "text/text")
}

func TestMimetypes_findMarshal(t *testing.T) {
	a := assert.New(t)
	mt := NewMimetypes()

	a.NotError(mt.Add("text", nil, nil))
	a.NotError(mt.Add("text/plain", nil, nil))
	a.NotError(mt.Add("text/text", nil, nil))
	a.NotError(mt.Add("application/aa", nil, nil)) // aa 排名靠前
	a.NotError(mt.Add("application/bb", nil, nil))

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
	a.NotError(mt.Add(DefaultMimetype, nil, nil))
	mm = mt.findMarshal("*/*")
	a.Equal(mm.name, DefaultMimetype)

	// 不存在
	a.Nil(mt.findMarshal("xx/*"))
}
