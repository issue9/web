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

func TestMimetypes_ContentType(t *testing.T) {
	a := assert.New(t)

	mt := NewMimetypes(DefaultBuilder)
	a.NotNil(mt)

	f, e, err := mt.ConentType(";;;")
	a.Error(err).Nil(f).Nil(e)

	// 不存在的 mimetype
	f, e, err = mt.ConentType(BuildContentType(DefaultMimetype, DefaultCharset))
	a.Error(err).Nil(f).Nil(e)

	mt.Add(DefaultMimetype, nil, json.Unmarshal)
	f, e, err = mt.ConentType(BuildContentType(DefaultMimetype, DefaultCharset))
	a.NotError(err).NotNil(f).NotNil(e)

	// 无效的字符集名称
	f, e, err = mt.ConentType(BuildContentType(DefaultMimetype, "invalid-charset"))
	a.Error(err).Nil(f).Nil(e)
}

func TestMimetypes_Unmarshal(t *testing.T) {
	a := assert.New(t)

	mt := NewMimetypes(DefaultBuilder)
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

	// 空的 unmarshal
	a.NotError(mt.Add("empty", json.Marshal, nil))
	um, err = mt.Unmarshal("empty")
	a.NotError(err).Nil(um)
}

func TestMimetypes_Marshal(t *testing.T) {
	a := assert.New(t)
	mt := NewMimetypes(DefaultBuilder)

	name, marshal, err := mt.Marshal(DefaultMimetype)
	a.Error(err).
		Nil(marshal).
		Empty(name)

	name, marshal, err = mt.Marshal("")
	a.ErrorIs(err, ErrNotFound).
		Nil(marshal).
		Empty(name)

	a.NotError(mt.Add(DefaultMimetype, xml.Marshal, xml.Unmarshal))
	a.NotError(mt.Add("text/plain", json.Marshal, json.Unmarshal))
	a.NotError(mt.Add("empty", nil, nil))

	name, marshal, err = mt.Marshal(DefaultMimetype)
	a.NotError(err).
		Equal(marshal, MarshalFunc(xml.Marshal)).
		Equal(name, DefaultMimetype)

	a.NotError(mt.Set(DefaultMimetype, json.Marshal, json.Unmarshal))
	name, marshal, err = mt.Marshal(DefaultMimetype)
	a.NotError(err).
		Equal(marshal, MarshalFunc(json.Marshal)).
		Equal(name, DefaultMimetype)

	a.ErrorIs(mt.Set("not-exists", nil, nil), ErrNotFound)

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

func TestMimetypes_Add_Delete(t *testing.T) {
	a := assert.New(t)
	mt := NewMimetypes(DefaultBuilder)
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

	// 删除
	mt.Delete("text")
	mt.Delete(DefaultMimetype)
	mt.Delete("not-exists")
	a.Equal(mt.codecs[0].name, "application/aa")
	a.Equal(mt.codecs[1].name, "application/bb")
	a.Equal(mt.codecs[2].name, "application/json")
	a.Equal(mt.codecs[3].name, "text/plain")
	a.Equal(mt.codecs[4].name, "text/text")
}

func TestMimetypes_findMarshal(t *testing.T) {
	a := assert.New(t)
	mt := NewMimetypes(DefaultBuilder)

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

func TestMimetypes_NewResult(t *testing.T) {
	a := assert.New(t)
	mgr := NewMimetypes(DefaultBuilder)
	mgr.AddMessage(400, 40000, "lang") // lang 有翻译

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

func TestMimetypes_NewResultWithFields(t *testing.T) {
	a := assert.New(t)
	mgr := NewMimetypes(DefaultBuilder)
	mgr.AddMessage(400, 40000, "lang") // lang 有翻译
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
	mgr := NewMimetypes(DefaultBuilder)

	a.NotPanic(func() {
		mgr.AddMessage(400, 1, "1")
		mgr.AddMessage(400, 100, "100")
	})

	msg, found := mgr.messages[1]
	a.True(found).
		Equal(msg.status, 400).
		Equal(msg.key, "1")

	msg, found = mgr.messages[401]
	a.False(found).Nil(msg)

	// 重复的 ID
	a.Panic(func() {
		mgr.AddMessage(400, 1, "40010")
	})
}

func TestMimetypes_Messages(t *testing.T) {
	a := assert.New(t)
	mgr := NewMimetypes(DefaultBuilder)
	a.NotNil(mgr)

	a.NotPanic(func() {
		mgr.AddMessage(400, 40010, "lang")
	})

	msg := mgr.Messages(message.NewPrinter(language.Und))
	a.Equal(msg[40010], "und")

	msg = mgr.Messages(message.NewPrinter(language.SimplifiedChinese))
	a.Equal(msg[40010], "hans")

	msg = mgr.Messages(message.NewPrinter(language.TraditionalChinese))
	a.Equal(msg[40010], "hant")

	msg = mgr.Messages(message.NewPrinter(language.English))
	a.Equal(msg[40010], "und")

	a.Panic(func() {
		mgr.Messages(nil)
	})
}
