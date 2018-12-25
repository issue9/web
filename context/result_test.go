// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
	"github.com/issue9/middleware/recovery"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	yaml "gopkg.in/yaml.v2"

	"github.com/issue9/web/mimetype/form"
)

var (
	_ form.Marshaler = &Result{}
	_ error          = &Result{}
)

func TestResult_Add_HasDetail(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)

	code := 400 * 1000
	a.NotPanic(func() { app.NewMessages(400, map[int]string{400000: "400"}) })
	r := &Result{Code: code}
	a.False(r.HasDetail())

	r.Add("field", "message")
	r.Add("field", "message")
	a.True(r.HasDetail())
	a.Equal(len(r.Detail), 2)
}

func TestResult_SetDetail(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)

	code := 400 * 1000
	a.NotPanic(func() { app.NewMessages(400, map[int]string{400000: "400"}) })
	r := &Result{Code: code}
	a.False(r.HasDetail())

	r.SetDetail(map[string]string{"field1": "message1", "field2": "message2"})
	r.SetDetail(map[string]string{"field1": "message1", "field2": "message2"})
	a.True(r.HasDetail())
	a.Equal(len(r.Detail), 2)
}

func TestResult_Render_Exit(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)

	a.NotPanic(func() {
		app.NewMessages(400, map[int]string{
			4000: "400", // 需要与 resultRenderHandler 中的错误代码值相同
			4001: "401", // 需要与 resultRenderHandler 中的错误代码值相同
		})
	})

	resultRenderHandler := func(w http.ResponseWriter, r *http.Request) {
		ctx := New(w, r, app)
		ctx.LocalePrinter = message.NewPrinter(language.Und)

		switch r.URL.Path {
		case "/render":
			rslt := ctx.NewResult(4000)
			rslt.SetDetail(map[string]string{"field1": "message1", "field2": "message2"})
			rslt.Render()
		case "/exit":
			rslt := ctx.NewResult(4001)
			rslt.SetDetail(map[string]string{"field1": "message1", "field2": "message2"})
			rslt.Exit()
		case "/error":
			rslt := &Result{Code: 100}
			rslt.Render()
		}
	}

	h := http.HandlerFunc(resultRenderHandler)
	srv := rest.NewServer(t, recovery.New(h, recoverFunc), nil)

	// render 的正常流程测试
	srv.NewRequest(http.MethodGet, "/render").
		Do().
		Status(400)

	// result.Code 不存在的情况
	srv.NewRequest(http.MethodGet, "/error").
		Do().
		Status(http.StatusInternalServerError) // 等同于 recoverFunc 中的输出报头

	// result.Exit() 测试
	srv.NewRequest(http.MethodGet, "/exit").
		Do().
		Status(400)
}

// 仅作为 TestResult_Render_Exit 的测试用
func recoverFunc(w http.ResponseWriter, msg interface{}) {
	w.WriteHeader(http.StatusInternalServerError)
}

var (
	mimetypeResult = &Result{
		Code:    400,
		Message: "400",
		Detail: []*detail{
			&detail{
				Field:   "field1",
				Message: "message1",
			},
			&detail{
				Field:   "field2",
				Message: "message2",
			},
		},
	}

	simpleMimetypeResult = &Result{
		Code:    400,
		Message: "400",
	}
)

func TestResultJSONMarshal(t *testing.T) {
	a := assert.New(t)

	bs, err := json.Marshal(mimetypeResult)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `{"message":"400","code":400,"detail":[{"field":"field1","message":"message1"},{"field":"field2","message":"message2"}]}`)

	bs, err = json.Marshal(simpleMimetypeResult)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `{"message":"400","code":400}`)
}

func TestResultXMLMarshal(t *testing.T) {
	a := assert.New(t)

	bs, err := xml.Marshal(mimetypeResult)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `<result message="400" code="400"><field name="field1">message1</field><field name="field2">message2</field></result>`)

	bs, err = xml.Marshal(simpleMimetypeResult)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `<result message="400" code="400"></result>`)
}

func TestResultYAMLMarshal(t *testing.T) {
	a := assert.New(t)

	bs, err := yaml.Marshal(mimetypeResult)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `message: "400"
code: 400
detail:
- field: field1
  message: message1
- field: field2
  message: message2
`)

	bs, err = yaml.Marshal(simpleMimetypeResult)
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `message: "400"
code: 400
`)
}

func TestResultFormMarshal(t *testing.T) {
	a := assert.New(t)

	bs, err := mimetypeResult.MarshalForm()
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `code=400&detail.field1=message1&detail.field2=message2&message=400`)

	bs, err = simpleMimetypeResult.MarshalForm()
	a.NotError(err).NotNil(bs)
	a.Equal(string(bs), `code=400&message=400`)
}