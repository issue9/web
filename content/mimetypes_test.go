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

	a.NotError(srv.Add(DefaultMimetype, json.Marshal, json.Unmarshal))

	um, err = srv.Unmarshal(DefaultMimetype)
	a.NotError(err).NotNil(um)

	// 未指定 mimetype
	um, err = srv.Unmarshal("")
	a.Error(err).Nil(um)

	// mimetype 无法找到
	um, err = srv.Unmarshal("not-exists")
	a.ErrorIs(err, ErrNotFound).Nil(um)
}

func TestMimetypes_Marshal(t *testing.T) {
	a := assert.New(t)
	srv := NewMimetypes()

	name, marshal, err := srv.Marshal(DefaultMimetype)
	a.Error(err).
		Nil(marshal).
		Empty(name)

	name, marshal, err = srv.Marshal("")
	a.ErrorIs(err, ErrNotFound).
		Nil(marshal).
		Empty(name)

	a.NotError(srv.Add(DefaultMimetype, json.Marshal, json.Unmarshal))
	a.NotError(srv.Add("text/plain", json.Marshal, json.Unmarshal))

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

func TestMimetypes_Add(t *testing.T) {
	a := assert.New(t)
	srv := NewMimetypes()
	a.NotNil(srv)

	// 不能添加同名的多次
	a.NotError(srv.Add(DefaultMimetype, nil, nil))
	a.ErrorIs(srv.Add(DefaultMimetype, nil, nil), ErrExists)

	// 不能添加以 /* 结属的名称
	a.Panic(func() {
		a.NotError(srv.Add("application/*", nil, nil))
	})
	a.Panic(func() {
		a.NotError(srv.Add("/*", nil, nil))
	})

	// 排序是否正常
	a.NotError(srv.Add("application/json", nil, nil))
	a.Equal(srv.codecs[0].name, DefaultMimetype) // 默认始终在第一
}

func TestMimetypes_findMarshal(t *testing.T) {
	a := assert.New(t)
	srv := NewMimetypes()

	a.NotError(srv.Add("text", nil, nil))
	a.NotError(srv.Add("text/plain", nil, nil))
	a.NotError(srv.Add("text/text", nil, nil))
	a.NotError(srv.Add("application/aa", nil, nil)) // aa 排名靠前
	a.NotError(srv.Add("application/bb", nil, nil))

	// 检测排序
	a.Equal(srv.codecs[0].name, "application/aa")
	a.Equal(srv.codecs[1].name, "application/bb")
	a.Equal(srv.codecs[2].name, "text")
	a.Equal(srv.codecs[3].name, "text/plain")
	a.Equal(srv.codecs[4].name, "text/text")

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
	a.NotError(srv.Add(DefaultMimetype, nil, nil))
	mm = srv.findMarshal("*/*")
	a.Equal(mm.name, DefaultMimetype)

	// 不存在
	a.Nil(srv.findMarshal("xx/*"))
}
