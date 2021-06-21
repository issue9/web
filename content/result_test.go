// SPDX-License-Identifier: MIT

package content

import (
	"encoding/json"
	"encoding/xml"
	"sort"
	"testing"

	"github.com/issue9/assert"
	"gopkg.in/yaml.v2"

	"github.com/issue9/web/content/form"
)

var _ BuildResultFunc = DefaultBuilder

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

	rslt := DefaultBuilder(500, 50001, "error message")
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

	rslt = DefaultBuilder(400, 40001, "400")
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
