// SPDX-License-Identifier: MIT

package context

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v2"

	"github.com/issue9/web/context/contentype/form"
)

var (
	_ BuildResultFunc = DefaultResultBuilder

	_ form.Marshaler   = &defaultResult{}
	_ form.Unmarshaler = &defaultResult{}
)

var (
	mimetypeResult = &defaultResult{
		Code:    400,
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
		Code:    400,
		Message: "400",
	}
)

func TestDefaultResult(t *testing.T) {
	a := assert.New(t)

	rslt := DefaultResultBuilder(500, 50001, "error message")
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

	rslt = DefaultResultBuilder(400, 40001, "400")
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
	a := assert.New(t)

	// marshal mimetypeResult
	bs, err := json.Marshal(mimetypeResult)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `{"message":"400","code":400,"fields":[{"name":"field1","message":["message1","message2"]},{"name":"field2","message":["message2"]}]}`)

	// unmarshal mimetypeResult
	obj := &defaultResult{}
	a.NotError(json.Unmarshal(bs, obj))
	a.Equal(obj, mimetypeResult)

	// marshal simpleMimetypesResult
	bs, err = json.Marshal(simpleMimetypeResult)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `{"message":"400","code":400}`)

	// unmarshal simpleMimetypesResult
	obj = &defaultResult{}
	a.NotError(json.Unmarshal(bs, obj))
	a.Equal(obj, simpleMimetypeResult)
}

func TestDefaultResultXML(t *testing.T) {
	a := assert.New(t)

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

func TestDefaultResultYAML(t *testing.T) {
	a := assert.New(t)

	// marshal mimetypeResult
	bs, err := yaml.Marshal(mimetypeResult)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `message: "400"
code: 400
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
code: 400
`)

	// unmarshal simpleMimetypesResult
	obj = &defaultResult{}
	a.NotError(yaml.Unmarshal(bs, obj))
	a.Equal(obj, simpleMimetypeResult)
}

func TestDefaultResultForm(t *testing.T) {
	a := assert.New(t)

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

func TestResult(t *testing.T) {
	a := assert.New(t)

	r := httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := newServer(a).NewContext(w, r)
	ctx.server.AddMessages(http.StatusBadRequest, map[int]message.Reference{
		40010: "40010",
		40011: "40011",
	})

	rslt := ctx.NewResultWithFields(40010, ResultFields{
		"k1": []string{"v1", "v2"},
	})
	a.True(rslt.HasFields())

	rslt.Render()
	a.Equal(w.Body.String(), `{"message":"40010","code":40010,"fields":[{"name":"k1","message":["v1","v2"]}]}`)
}

func TestContext_NewResult(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)
	srv.AddMessages(400, map[int]message.Reference{40000: "lang"}) // lang 有翻译
	w := httptest.NewRecorder()

	// 能正常翻译错误信息
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("accept-language", language.SimplifiedChinese.String())
	ctx := srv.NewContext(w, r)
	rslt, ok := ctx.NewResult(40000).rslt.(*defaultResult)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "hans")

	// 未指定 accept-language，采用默认的 und
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	ctx = srv.NewContext(w, r)
	rslt, ok = ctx.NewResult(40000).rslt.(*defaultResult)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "und")

	// 不存在的本地化信息，采用默认的 und
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("accept-language", "en-US")
	ctx = srv.NewContext(w, r)
	rslt, ok = ctx.NewResult(40000).rslt.(*defaultResult)
	a.True(ok).NotNil(rslt)
	a.Equal(rslt.Message, "und")

	// 不存在
	a.Panic(func() { ctx.NewResult(400) })
	a.Panic(func() { ctx.NewResult(50000) })
}

func TestServer_AddMessages(t *testing.T) {
	a := assert.New(t)
	builder := newServer(a)

	a.NotPanic(func() {
		builder.AddMessages(400, map[int]message.Reference{
			1:   "1",
			100: "100",
		})
	})

	msg, found := builder.messages[1]
	a.True(found).
		Equal(msg.status, 400).
		Equal(msg.message, "1")

	msg, found = builder.messages[401]
	a.False(found).Nil(msg)

	// 消息不能为空
	a.Panic(func() {
		builder.AddMessages(400, map[int]message.Reference{
			1:   "",
			100: "100",
		})
	})

	// 重复的 ID
	a.Panic(func() {
		builder.AddMessages(400, map[int]message.Reference{
			1:   "1",
			100: "100",
		})
	})
}

func TestServer_Messages(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)
	a.NotNil(srv)

	a.NotPanic(func() {
		srv.AddMessages(400, map[int]message.Reference{40010: "lang"})
	})

	lmsgs := srv.Messages(message.NewPrinter(language.Und, message.Catalog(srv.Catalog)))
	a.Equal(lmsgs[40010], "und")

	lmsgs = srv.Messages(message.NewPrinter(language.SimplifiedChinese, message.Catalog(srv.Catalog)))
	a.Equal(lmsgs[40010], "hans")

	lmsgs = srv.Messages(message.NewPrinter(language.TraditionalChinese, message.Catalog(srv.Catalog)))
	a.Equal(lmsgs[40010], "hant")

	lmsgs = srv.Messages(message.NewPrinter(language.English, message.Catalog(srv.Catalog)))
	a.Equal(lmsgs[40010], "und")

	lmsgs = srv.Messages(nil)
	a.Equal(lmsgs[40010], "und")
}
