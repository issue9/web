// SPDX-License-Identifier: MIT

package server

import (
	"encoding/json"
	"encoding/xml"
	"sort"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/localeutil"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"

	"github.com/issue9/web/serializer/form"
	"github.com/issue9/web/serializer/protobuf"
	"github.com/issue9/web/server/testdata"
)

var (
	_ BuildErrInfoFunc = DefaultErrInfoBuilder
	_ form.Marshaler   = &errInfo{}
	_ form.Unmarshaler = &errInfo{}
	_ proto.Message    = &errInfo{}
)

var (
	jsonErrInfo = &errInfo{
		Code:    "400",
		Message: "400",
		Fields: []*errField{
			{
				Name:    "field1",
				Message: []string{"message1", "message2"},
			},
			{
				Name:    "field2",
				Message: []string{"message2"},
			},
		},
	}

	simpleJSONErrInfo = &errInfo{
		Code:    "400",
		Message: "400",
	}
)

func TestDefaultErrInfo_JSON(t *testing.T) {
	a := assert.New(t, false)

	// marshal jsonErrInfo
	bs, err := json.Marshal(jsonErrInfo)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `{"message":"400","code":"400","fields":[{"name":"field1","message":["message1","message2"]},{"name":"field2","message":["message2"]}]}`)

	// unmarshal jsonErrInfo
	obj := &errInfo{}
	a.NotError(json.Unmarshal(bs, obj))
	a.Equal(obj, jsonErrInfo)

	// marshal simpleJSONErrInfo
	bs, err = json.Marshal(simpleJSONErrInfo)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `{"message":"400","code":"400"}`)

	// unmarshal simpleJSONErrInfo
	obj = &errInfo{}
	a.NotError(json.Unmarshal(bs, obj))
	a.Equal(obj, simpleJSONErrInfo)
}

func TestDefaultErrInfo_XML(t *testing.T) {
	a := assert.New(t, false)

	// marshal jsonErrInfo
	bs, err := xml.Marshal(jsonErrInfo)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `<errors code="400"><message>400</message><field name="field1"><message>message1</message><message>message2</message></field><field name="field2"><message>message2</message></field></errors>`)

	// unmarshal jsonErrInfo
	obj := &errInfo{}
	a.NotError(xml.Unmarshal(bs, obj))
	a.Equal(obj, jsonErrInfo)

	// marshal simpleJSONErrInfo
	bs, err = xml.Marshal(simpleJSONErrInfo)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `<errors code="400"><message>400</message></errors>`)

	// unmarshal simpleJSONErrInfo
	obj = &errInfo{}
	a.NotError(xml.Unmarshal(bs, obj))
	a.Equal(obj, simpleJSONErrInfo)
}

func TestDefaultErrInfo_Protobuf(t *testing.T) {
	a := assert.New(t, false)

	// marshal jsonErrInfo
	bs, err := protobuf.Marshal(jsonErrInfo)
	a.NotError(err).NotNil(bs)

	// unmarshal jsonErrInfo
	obj := &testdata.Errors{}
	a.NotError(protobuf.Unmarshal(bs, obj))
	a.Equal(obj.Message, jsonErrInfo.Message).
		Equal(obj.Code, jsonErrInfo.Code).
		Equal(2, len(obj.Fields)).
		Equal(obj.Fields[0].Name, "field1").
		Equal(obj.Fields[0].Message, []string{"message1", "message2"})

	// marshal simpleJSONErrInfo
	bs, err = protobuf.Marshal(simpleJSONErrInfo)
	a.NotError(err).NotNil(bs)

	// unmarshal simpleJSONErrInfo
	obj = &testdata.Errors{}
	a.NotError(protobuf.Unmarshal(bs, obj))
	a.Equal(obj.Message, simpleJSONErrInfo.Message).
		Equal(obj.Code, simpleJSONErrInfo.Code)
}

func TestDefaultErrInfo_YAML(t *testing.T) {
	a := assert.New(t, false)

	// marshal jsonErrInfo
	bs, err := yaml.Marshal(jsonErrInfo)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `message: "400"
code: "400"
fields:
    - name: field1
      message:
        - message1
        - message2
    - name: field2
      message:
        - message2
`)

	// unmarshal jsonErrInfo
	obj := &errInfo{}
	a.NotError(yaml.Unmarshal(bs, obj))
	a.Equal(obj, jsonErrInfo)

	// marshal simpleJSONErrInfo
	bs, err = yaml.Marshal(simpleJSONErrInfo)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `message: "400"
code: "400"
`)

	// unmarshal simpleJSONErrInfo
	obj = &errInfo{}
	a.NotError(yaml.Unmarshal(bs, obj))
	a.Equal(obj, simpleJSONErrInfo)
}

func TestDefaultErrInfo_Form(t *testing.T) {
	a := assert.New(t, false)

	// marshal jsonErrInfo
	bs, err := form.Marshal(jsonErrInfo)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `code=400&fields.field1=message1&fields.field1=message2&fields.field2=message2&message=400`)

	// unmarshal jsonErrInfo
	obj := &errInfo{}
	a.NotError(form.Unmarshal(bs, obj))
	sort.SliceStable(obj.Fields, func(i, j int) bool { return obj.Fields[i].Name < obj.Fields[j].Name }) // 顺序一致才能相等
	a.Equal(obj, jsonErrInfo)

	// marshal simpleJSONErrInfo
	bs, err = form.Marshal(simpleJSONErrInfo)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `code=400&message=400`)

	// unmarshal simpleJSONErrInfo
	obj = &errInfo{}
	a.NotError(form.Unmarshal(bs, obj))
	a.Equal(obj, simpleJSONErrInfo)
}

func TestServer_ErrInfo(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)

	srv.AddErrInfo(400, "40000", localeutil.Phrase("lang")) // lang 有翻译

	// 能正常翻译错误信息
	rslt, ok := srv.ErrInfo(srv.NewPrinter(language.SimplifiedChinese), "40000", nil).(*errInfo)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "hans")

	// 采用 und
	rslt, ok = srv.ErrInfo(srv.NewPrinter(language.Und), "40000", nil).(*errInfo)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "und")

	// 不存在的本地化信息，采用默认的 und
	rslt, ok = srv.ErrInfo(srv.NewPrinter(language.Afrikaans), "40000", nil).(*errInfo)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "und")

	// 不存在
	a.Panic(func() { srv.ErrInfo(srv.NewPrinter(language.Afrikaans), "400", nil) })
	a.Panic(func() { srv.ErrInfo(srv.NewPrinter(language.Afrikaans), "50000", nil) })

	// with fields

	fields := map[string][]string{"f1": {"v1", "v2"}}

	// 能正常翻译错误信息
	rslt, ok = srv.ErrInfo(srv.NewPrinter(language.SimplifiedChinese), "40000", fields).(*errInfo)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "hans").
		Equal(rslt.Fields, []*errField{{Name: "f1", Message: []string{"v1", "v2"}}})

	// 采用 und
	rslt, ok = srv.ErrInfo(srv.NewPrinter(language.Und), "40000", fields).(*errInfo)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "und").
		Equal(rslt.Fields, []*errField{{Name: "f1", Message: []string{"v1", "v2"}}})
}

func TestServer_AddErrInfo(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, &Options{LanguageTag: language.SimplifiedChinese})

	a.NotPanic(func() {
		srv.AddErrInfo(400, "1", localeutil.Phrase("1"))
		srv.AddErrInfo(400, "100", localeutil.Phrase("100"))
	})

	msg, found := srv.errInfo["1"]
	a.True(found).
		Equal(msg.status, 400)

	msg, found = srv.errInfo["401"]
	a.False(found).Nil(msg)

	// 重复的 ID
	a.Panic(func() {
		srv.AddErrInfo(400, "1", localeutil.Phrase("40010"))
	})
}

func TestServer_ErrInfos(t *testing.T) {
	a := assert.New(t, false)
	c := newServer(a, &Options{LanguageTag: language.SimplifiedChinese})

	a.NotPanic(func() {
		c.AddErrInfos(400, map[string]localeutil.LocaleStringer{"40010": localeutil.Phrase("lang")})
	})

	msg := c.ErrInfos(c.NewPrinter(language.Und))
	a.Equal(msg["40010"], "und")

	msg = c.ErrInfos(c.NewPrinter(language.SimplifiedChinese))
	a.Equal(msg["40010"], "hans")

	msg = c.ErrInfos(c.NewPrinter(language.TraditionalChinese))
	a.Equal(msg["40010"], "hant")

	msg = c.ErrInfos(c.NewPrinter(language.English))
	a.Equal(msg["40010"], "und")

	a.Panic(func() {
		c.ErrInfos(nil)
	})
}
