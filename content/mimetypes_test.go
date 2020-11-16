// SPDX-License-Identifier: MIT

package content

import (
	"encoding/json"
	"testing"

	"github.com/issue9/assert"
)

func TestMimetypes_Unmarshal(t *testing.T) {
	a := assert.New(t)

	srv := NewMimetypes()
	a.NotNil(srv)

	um, err := srv.Unmarshal("")
	a.Error(err).
		Nil(um)

	a.NotError(srv.AddUnmarshal(DefaultMimetype, json.Unmarshal))
	a.NotError(srv.AddMarshal(DefaultMimetype, json.Marshal))

	// 未指定 mimetype
	um, err = srv.Unmarshal("")
	a.Error(err).Nil(um)

	// mimetype 无法找到
	um, err = srv.Unmarshal("not-exists")
	a.Error(err).Nil(um)
}

func TestMimetypes_Marshal(t *testing.T) {
	a := assert.New(t)
	srv := NewMimetypes()

	name, marshal, err := srv.Marshal(DefaultMimetype)
	a.Error(err).
		Nil(marshal).
		Empty(name)

	name, marshal, err = srv.Marshal("")
	a.ErrorString(err, "请求中未指定 accept 报头，且服务端也未指定匹配 */* 的解码函数").
		Nil(marshal).
		Empty(name)

	a.NotError(srv.AddMarshal(DefaultMimetype, json.Marshal))
	a.NotError(srv.AddMarshal("text/plain", json.Marshal))

	name, marshal, err = srv.Marshal(DefaultMimetype)
	a.NotError(err).
		Equal(marshal, MarshalFunc(json.Marshal)).
		Equal(name, DefaultMimetype)

	name, marshal, err = srv.Marshal(DefaultMimetype)
	a.NotError(err).
		Equal(marshal, MarshalFunc(json.Marshal)).
		Equal(name, DefaultMimetype)

	// */* 如果指定了 DefaultMimetype，则必定是该值
	name, marshal, err = srv.Marshal("*/*")
	a.NotError(err).
		Equal(marshal, MarshalFunc(json.Marshal)).
		Equal(name, DefaultMimetype)

	// 同 */*
	name, marshal, err = srv.Marshal("")
	a.NotError(err).
		Equal(marshal, MarshalFunc(json.Marshal)).
		Equal(name, DefaultMimetype)

	name, marshal, err = srv.Marshal("*/*,text/plain")
	a.NotError(err).
		Equal(marshal, MarshalFunc(json.Marshal)).
		Equal(name, "text/plain")

	name, marshal, err = srv.Marshal("font/wottf;q=x.9")
	a.Error(err).
		Empty(name).
		Nil(marshal)

	name, marshal, err = srv.Marshal("font/wottf")
	a.Error(err).
		Empty(name).
		Nil(marshal)
}

func TestMimetypes_AddMarshal(t *testing.T) {
	a := assert.New(t)
	srv := NewMimetypes()
	a.NotNil(srv)

	// 不能添加同名的多次
	a.NotError(srv.AddMarshal(DefaultMimetype, nil))
	a.Error(srv.AddMarshal(DefaultMimetype, nil))

	// 不能添加以 /* 结属的名称
	a.Panic(func() {
		a.NotError(srv.AddMarshal("application/*", nil))
	})
	a.Panic(func() {
		a.NotError(srv.AddMarshal("/*", nil))
	})

	// 排序是否正常
	a.NotError(srv.AddMarshal("application/json", nil))
	a.Equal(srv.marshals[0].name, DefaultMimetype) // 默认始终在第一
}

func TestMimetypes_AddUnmarshal(t *testing.T) {
	a := assert.New(t)
	srv := NewMimetypes()
	a.NotNil(srv)

	a.NotError(srv.AddUnmarshal(DefaultMimetype, nil))
	a.Error(srv.AddUnmarshal(DefaultMimetype, nil))

	// 不能添加包含 * 字符的名称
	a.Panic(func() {
		a.NotError(srv.AddUnmarshal("application/*", nil))
	})
	a.Panic(func() {
		a.NotError(srv.AddUnmarshal("*", nil))
	})

	// 排序是否正常
	a.NotError(srv.AddUnmarshal("application/json", nil))
	a.Equal(srv.unmarshals[0].name, DefaultMimetype) // 默认始终在第一
}

func TestMimetypes_AddUnmarshals(t *testing.T) {
	a := assert.New(t)
	srv := NewMimetypes()
	a.NotNil(srv)

	err := srv.AddUnmarshals(map[string]UnmarshalFunc{
		DefaultMimetype:    nil,
		"text":             nil,
		"application/json": nil,
		"application/xml":  nil,
	})
	a.NotError(err)

	a.Equal(srv.unmarshals[0].name, DefaultMimetype)
	a.Equal(srv.unmarshals[1].name, "application/json")
	a.Equal(srv.unmarshals[2].name, "application/xml")
	a.Equal(srv.unmarshals[3].name, "text")

	_, err = srv.Unmarshal("*/*")
	a.ErrorString(err, "未找到 */* 类型的解码函数")

	_, err = srv.Unmarshal("text")
	a.NotError(err)
}

func TestMimetypes_findMarshal(t *testing.T) {
	a := assert.New(t)
	srv := NewMimetypes()

	a.NotError(srv.AddMarshals(map[string]MarshalFunc{
		"text":           nil,
		"text/plain":     nil,
		"text/text":      nil,
		"application/aa": nil, // aa 排名靠前
		"application/bb": nil, // aa 排名靠前
	}))

	// 检测排序
	a.Equal(srv.marshals[0].name, "application/aa")
	a.Equal(srv.marshals[1].name, "application/bb")
	a.Equal(srv.marshals[2].name, "text")
	a.Equal(srv.marshals[3].name, "text/plain")
	a.Equal(srv.marshals[4].name, "text/text")

	mm := srv.findMarshal("text")
	a.Equal(mm.name, "text")

	mm = srv.findMarshal("text/*")
	a.Equal(mm.name, "text")

	mm = srv.findMarshal("application/*")
	a.Equal(mm.name, "application/aa")

	// 第一条数据
	mm = srv.findMarshal("*/*")
	a.Equal(mm.name, "application/aa")

	// 第一条数据
	mm = srv.findMarshal("")
	a.Equal(mm.name, "application/aa")

	// 有默认值，则始终在第一
	a.NotError(srv.AddMarshal(DefaultMimetype, nil))
	mm = srv.findMarshal("*/*")
	a.Equal(mm.name, DefaultMimetype)

	// 不存在
	a.Nil(srv.findMarshal("xx/*"))
}
