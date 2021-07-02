// SPDX-License-Identifier: MIT

package content

import (
	"encoding/json"
	"encoding/xml"
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func init() {
	if err := message.SetString(language.Und, "lang", "und"); err != nil {
		panic(err)
	}

	if err := message.SetString(language.SimplifiedChinese, "lang", "hans"); err != nil {
		panic(err)
	}

	if err := message.SetString(language.TraditionalChinese, "lang", "hant"); err != nil {
		panic(err)
	}
}

func TestContent_ContentType(t *testing.T) {
	a := assert.New(t)

	mt := New(DefaultBuilder)
	a.NotNil(mt)

	f, e, err := mt.ConentType(";;;")
	a.Error(err).Nil(f).Nil(e)

	// 不存在的 mimetype
	f, e, err = mt.ConentType(BuildContentType(DefaultMimetype, DefaultCharset))
	a.Error(err).Nil(f).Nil(e)

	mt.AddMimetype(DefaultMimetype, nil, json.Unmarshal)
	f, e, err = mt.ConentType(BuildContentType(DefaultMimetype, DefaultCharset))
	a.NotError(err).NotNil(f).NotNil(e)

	// 无效的字符集名称
	f, e, err = mt.ConentType(BuildContentType(DefaultMimetype, "invalid-charset"))
	a.Error(err).Nil(f).Nil(e)
}

func TestContent_Unmarshal(t *testing.T) {
	a := assert.New(t)

	mt := New(DefaultBuilder)
	a.NotNil(mt)

	um, err := mt.Unmarshal("")
	a.Error(err).
		Nil(um)

	a.NotError(mt.AddMimetype(DefaultMimetype, json.Marshal, json.Unmarshal))

	um, err = mt.Unmarshal(DefaultMimetype)
	a.NotError(err).NotNil(um)

	// 未指定 mimetype
	um, err = mt.Unmarshal("")
	a.Error(err).Nil(um)

	// mimetype 无法找到
	um, err = mt.Unmarshal("not-exists")
	a.ErrorIs(err, ErrNotFound).Nil(um)

	// 空的 unmarshal
	a.NotError(mt.AddMimetype("empty", json.Marshal, nil))
	um, err = mt.Unmarshal("empty")
	a.NotError(err).Nil(um)
}

func TestContent_Marshal(t *testing.T) {
	a := assert.New(t)
	mt := New(DefaultBuilder)

	name, marshal, err := mt.Marshal(DefaultMimetype)
	a.Error(err).
		Nil(marshal).
		Empty(name)

	name, marshal, err = mt.Marshal("")
	a.ErrorIs(err, ErrNotFound).
		Nil(marshal).
		Empty(name)

	a.NotError(mt.AddMimetype(DefaultMimetype, xml.Marshal, xml.Unmarshal))
	a.NotError(mt.AddMimetype("text/plain", json.Marshal, json.Unmarshal))
	a.NotError(mt.AddMimetype("empty", nil, nil))

	name, marshal, err = mt.Marshal(DefaultMimetype)
	a.NotError(err).
		Equal(marshal, MarshalFunc(xml.Marshal)).
		Equal(name, DefaultMimetype)

	a.NotError(mt.SetMimetype(DefaultMimetype, json.Marshal, json.Unmarshal))
	name, marshal, err = mt.Marshal(DefaultMimetype)
	a.NotError(err).
		Equal(marshal, MarshalFunc(json.Marshal)).
		Equal(name, DefaultMimetype)

	a.ErrorIs(mt.SetMimetype("not-exists", nil, nil), ErrNotFound)

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

	// 匹配 empty
	name, marshal, err = mt.Marshal("empty")
	a.NotError(err).
		Equal(name, "empty").
		Nil(marshal)
}

func TestContent_Add_Delete(t *testing.T) {
	a := assert.New(t)
	mt := New(DefaultBuilder)
	a.NotNil(mt)

	// 不能添加同名的多次
	a.NotError(mt.AddMimetype(DefaultMimetype, nil, nil))
	a.ErrorIs(mt.AddMimetype(DefaultMimetype, nil, nil), ErrExists)

	// 不能添加以 /* 结属的名称
	a.Panic(func() {
		a.NotError(mt.AddMimetype("application/*", nil, nil))
	})
	a.Panic(func() {
		a.NotError(mt.AddMimetype("/*", nil, nil))
	})

	// 排序是否正常
	a.NotError(mt.AddMimetype("application/json", nil, nil))
	a.Equal(mt.mimetypes[0].name, DefaultMimetype) // 默认始终在第一

	a.NotError(mt.AddMimetype("text", nil, nil))
	a.NotError(mt.AddMimetype("text/plain", nil, nil))
	a.NotError(mt.AddMimetype("text/text", nil, nil))
	a.NotError(mt.AddMimetype("application/aa", nil, nil)) // aa 排名靠前
	a.NotError(mt.AddMimetype("application/bb", nil, nil))

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

	a.NotError(mt.AddMimetype("text", nil, nil))
	a.NotError(mt.AddMimetype("text/plain", nil, nil))
	a.NotError(mt.AddMimetype("text/text", nil, nil))
	a.NotError(mt.AddMimetype("application/aa", nil, nil)) // aa 排名靠前
	a.NotError(mt.AddMimetype("application/bb", nil, nil))

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
	a.NotError(mt.AddMimetype(DefaultMimetype, nil, nil))
	mm = mt.findMarshal("*/*")
	a.Equal(mm.name, DefaultMimetype)

	// 不存在
	a.Nil(mt.findMarshal("xx/*"))
}

func TestContent_NewResult(t *testing.T) {
	a := assert.New(t)
	mgr := New(DefaultBuilder)
	mgr.AddResult(400, 40000, "lang") // lang 有翻译

	// 能正常翻译错误信息
	rslt, ok := mgr.NewResult(message.NewPrinter(language.SimplifiedChinese), 40000).(*defaultResult)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "hans")

	// 采用 und
	rslt, ok = mgr.NewResult(message.NewPrinter(language.Und), 40000).(*defaultResult)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "und")

	// 不存在的本地化信息，采用默认的 und
	rslt, ok = mgr.NewResult(message.NewPrinter(language.Afrikaans), 40000).(*defaultResult)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "und")

	// 不存在
	a.Panic(func() { mgr.NewResult(message.NewPrinter(language.Afrikaans), 400) })
	a.Panic(func() { mgr.NewResult(message.NewPrinter(language.Afrikaans), 50000) })
}

func TestContent_NewResultWithFields(t *testing.T) {
	a := assert.New(t)
	mgr := New(DefaultBuilder)
	mgr.AddResult(400, 40000, "lang") // lang 有翻译
	fields := map[string][]string{"f1": {"v1", "v2"}}

	// 能正常翻译错误信息
	rslt, ok := mgr.NewResultWithFields(message.NewPrinter(language.SimplifiedChinese), 40000, fields).(*defaultResult)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "hans").
		Equal(rslt.Fields, []*fieldDetail{{Name: "f1", Message: []string{"v1", "v2"}}})

	// 采用 und
	rslt, ok = mgr.NewResultWithFields(message.NewPrinter(language.Und), 40000, fields).(*defaultResult)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "und").
		Equal(rslt.Fields, []*fieldDetail{{Name: "f1", Message: []string{"v1", "v2"}}})
}

func TestMinetypes_AddMessage(t *testing.T) {
	a := assert.New(t)
	mgr := New(DefaultBuilder)

	a.NotPanic(func() {
		mgr.AddResult(400, 1, "1")
		mgr.AddResult(400, 100, "100")
	})

	msg, found := mgr.resultMessages[1]
	a.True(found).
		Equal(msg.status, 400).
		Equal(msg.key, "1")

	msg, found = mgr.resultMessages[401]
	a.False(found).Nil(msg)

	// 重复的 ID
	a.Panic(func() {
		mgr.AddResult(400, 1, "40010")
	})
}

func TestContent_Messages(t *testing.T) {
	a := assert.New(t)
	mgr := New(DefaultBuilder)
	a.NotNil(mgr)

	a.NotPanic(func() {
		mgr.AddResult(400, 40010, "lang")
	})

	msg := mgr.Results(message.NewPrinter(language.Und))
	a.Equal(msg[40010], "und")

	msg = mgr.Results(message.NewPrinter(language.SimplifiedChinese))
	a.Equal(msg[40010], "hans")

	msg = mgr.Results(message.NewPrinter(language.TraditionalChinese))
	a.Equal(msg[40010], "hant")

	msg = mgr.Results(message.NewPrinter(language.English))
	a.Equal(msg[40010], "und")

	a.Panic(func() {
		mgr.Results(nil)
	})
}
