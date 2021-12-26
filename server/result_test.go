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
	"gopkg.in/yaml.v2"

	"github.com/issue9/web/serialization/form"
	"github.com/issue9/web/server/testdata"
)

var (
	_ BuildResultFunc  = DefaultResultBuilder
	_ form.Marshaler   = &defaultResult{}
	_ form.Unmarshaler = &defaultResult{}
	_ proto.Message    = &defaultResult{}
)

var (
	mimetypeResult = &defaultResult{
		Code:    "400",
		Message: "400",
		Fields: []*fieldDetail{
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

	simpleMimetypeResult = &defaultResult{
		Code:    "400",
		Message: "400",
	}
)

func TestDefaultResult(t *testing.T) {
	a := assert.New(t, false)

	rslt := DefaultResultBuilder(500, "50001", "error message")
	a.False(rslt.HasFields()).
		Equal(rslt.Status(), 500)

	rslt.Add("f1", "f1 msg1")
	rslt.Add("f1", "f1 msg2")
	a.True(rslt.HasFields())
	r, ok := rslt.(*defaultResult)
	a.True(ok).Equal(2, len(r.Fields[0].Message))

	rslt.Set("f1", "f1 msg")
	a.True(rslt.HasFields())
	r, ok = rslt.(*defaultResult)
	a.True(ok).Equal(1, len(r.Fields[0].Message))

	rslt = DefaultResultBuilder(400, "40001", "400")
	rslt.Set("f1", "f1 msg1")
	a.True(rslt.HasFields())
	r, ok = rslt.(*defaultResult)
	a.True(ok).
		Equal(1, len(r.Fields[0].Message)).
		Equal("f1 msg1", r.Fields[0].Message[0])

	rslt.Set("f1", "f1 msg2")
	a.True(rslt.HasFields())
	r, ok = rslt.(*defaultResult)
	a.True(ok).
		Equal(1, len(r.Fields[0].Message)).
		Equal("f1 msg2", r.Fields[0].Message[0])
}

func TestDefaultResultJSON(t *testing.T) {
	a := assert.New(t, false)

	// marshal mimetypeResult
	bs, err := json.Marshal(mimetypeResult)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `{"message":"400","code":"400","fields":[{"name":"field1","message":["message1","message2"]},{"name":"field2","message":["message2"]}]}`)

	// unmarshal mimetypeResult
	obj := &defaultResult{}
	a.NotError(json.Unmarshal(bs, obj))
	a.Equal(obj, mimetypeResult)

	// marshal simpleMimetypesResult
	bs, err = json.Marshal(simpleMimetypeResult)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `{"message":"400","code":"400"}`)

	// unmarshal simpleMimetypesResult
	obj = &defaultResult{}
	a.NotError(json.Unmarshal(bs, obj))
	a.Equal(obj, simpleMimetypeResult)
}

func TestDefaultResultXML(t *testing.T) {
	a := assert.New(t, false)

	// marshal mimetypeResult
	bs, err := xml.Marshal(mimetypeResult)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `<result code="400"><message>400</message><field name="field1"><message>message1</message><message>message2</message></field><field name="field2"><message>message2</message></field></result>`)

	// unmarshal mimetypeResult
	obj := &defaultResult{}
	a.NotError(xml.Unmarshal(bs, obj))
	a.Equal(obj, mimetypeResult)

	// marshal simpleMimetypesResult
	bs, err = xml.Marshal(simpleMimetypeResult)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `<result code="400"><message>400</message></result>`)

	// unmarshal simpleMimetypesResult
	obj = &defaultResult{}
	a.NotError(xml.Unmarshal(bs, obj))
	a.Equal(obj, simpleMimetypeResult)
}

func TestDefaultResultProtobuf(t *testing.T) {
	a := assert.New(t, false)

	// marshal mimetypeResult
	bs, err := proto.Marshal(mimetypeResult)
	a.NotError(err).NotNil(bs)

	// unmarshal mimetypeResult
	obj := &testdata.Result{}
	a.NotError(proto.Unmarshal(bs, obj))
	a.Equal(obj.Message, mimetypeResult.Message).
		Equal(obj.Code, mimetypeResult.Code).
		Equal(2, len(obj.Fields)).
		Equal(obj.Fields[0].Name, "field1").
		Equal(obj.Fields[0].Message, []string{"message1", "message2"})

	// marshal simpleMimetypesResult
	bs, err = proto.Marshal(simpleMimetypeResult)
	a.NotError(err).NotNil(bs)

	// unmarshal simpleMimetypesResult
	obj = &testdata.Result{}
	a.NotError(proto.Unmarshal(bs, obj))
	a.Equal(obj.Message, simpleMimetypeResult.Message).
		Equal(obj.Code, simpleMimetypeResult.Code)
}

func TestDefaultResultYAML(t *testing.T) {
	a := assert.New(t, false)

	// marshal mimetypeResult
	bs, err := yaml.Marshal(mimetypeResult)
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

	// unmarshal mimetypeResult
	obj := &defaultResult{}
	a.NotError(yaml.Unmarshal(bs, obj))
	a.Equal(obj, mimetypeResult)

	// marshal simpleMimetypesResult
	bs, err = yaml.Marshal(simpleMimetypeResult)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `message: "400"
code: "400"
`)

	// unmarshal simpleMimetypesResult
	obj = &defaultResult{}
	a.NotError(yaml.Unmarshal(bs, obj))
	a.Equal(obj, simpleMimetypeResult)
}

func TestDefaultResultForm(t *testing.T) {
	a := assert.New(t, false)

	// marshal mimetypeResult
	bs, err := form.Marshal(mimetypeResult)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `code=400&fields.field1=message1&fields.field1=message2&fields.field2=message2&message=400`)

	// unmarshal mimetypeResult
	obj := &defaultResult{}
	a.NotError(form.Unmarshal(bs, obj))
	sort.SliceStable(obj.Fields, func(i, j int) bool { return obj.Fields[i].Name < obj.Fields[j].Name }) // 顺序一致才能相等
	a.Equal(obj, mimetypeResult)

	// marshal simpleMimetypesResult
	bs, err = form.Marshal(simpleMimetypeResult)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `code=400&message=400`)

	// unmarshal simpleMimetypesResult
	obj = &defaultResult{}
	a.NotError(form.Unmarshal(bs, obj))
	a.Equal(obj, simpleMimetypeResult)
}

func TestServer_Result(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, nil)

	srv.AddResult(400, "40000", localeutil.Phrase("lang")) // lang 有翻译

	// 能正常翻译错误信息
	rslt, ok := srv.Result(srv.Locale().Printer(language.SimplifiedChinese), "40000", nil).(*defaultResult)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "hans")

	// 采用 und
	rslt, ok = srv.Result(srv.Locale().Printer(language.Und), "40000", nil).(*defaultResult)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "und")

	// 不存在的本地化信息，采用默认的 und
	rslt, ok = srv.Result(srv.Locale().Printer(language.Afrikaans), "40000", nil).(*defaultResult)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "und")

	// 不存在
	a.Panic(func() { srv.Result(srv.Locale().Printer(language.Afrikaans), "400", nil) })
	a.Panic(func() { srv.Result(srv.Locale().Printer(language.Afrikaans), "50000", nil) })

	// with fields

	fields := map[string][]string{"f1": {"v1", "v2"}}

	// 能正常翻译错误信息
	rslt, ok = srv.Result(srv.Locale().Printer(language.SimplifiedChinese), "40000", fields).(*defaultResult)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "hans").
		Equal(rslt.Fields, []*fieldDetail{{Name: "f1", Message: []string{"v1", "v2"}}})

	// 采用 und
	rslt, ok = srv.Result(srv.Locale().Printer(language.Und), "40000", fields).(*defaultResult)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "und").
		Equal(rslt.Fields, []*fieldDetail{{Name: "f1", Message: []string{"v1", "v2"}}})
}

func TestServer_AddResult(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a, &Options{Tag: language.SimplifiedChinese})

	a.NotPanic(func() {
		srv.AddResult(400, "1", localeutil.Phrase("1"))
		srv.AddResult(400, "100", localeutil.Phrase("100"))
	})

	msg, found := srv.resultMessages["1"]
	a.True(found).
		Equal(msg.status, 400)

	msg, found = srv.resultMessages["401"]
	a.False(found).Nil(msg)

	// 重复的 ID
	a.Panic(func() {
		srv.AddResult(400, "1", localeutil.Phrase("40010"))
	})
}

func TestServer_Results(t *testing.T) {
	a := assert.New(t, false)
	c := newServer(a, &Options{Tag: language.SimplifiedChinese})

	a.NotPanic(func() {
		c.AddResults(400, map[string]localeutil.LocaleStringer{"40010": localeutil.Phrase("lang")})
	})

	msg := c.Results(c.Locale().Printer(language.Und))
	a.Equal(msg["40010"], "und")

	msg = c.Results(c.Locale().Printer(language.SimplifiedChinese))
	a.Equal(msg["40010"], "hans")

	msg = c.Results(c.Locale().Printer(language.TraditionalChinese))
	a.Equal(msg["40010"], "hant")

	msg = c.Results(c.Locale().Printer(language.English))
	a.Equal(msg["40010"], "und")

	a.Panic(func() {
		c.Results(nil)
	})
}
