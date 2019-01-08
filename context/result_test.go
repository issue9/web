// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
	yaml "gopkg.in/yaml.v2"

	"github.com/issue9/web/mimetype/form"
)

var (
	_ form.Marshaler = &Result{}
	_ error          = &Result{}
)

func TestContext_NewResult(t *testing.T) {
	a := assert.New(t)
	app := newApp(a)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	ctx := New(w, r, app)

	// 不存在
	a.Panic(func() { ctx.NewResult(400) })

	a.NotPanic(func() { app.NewMessages(400, map[int]string{40000: "400"}) })
	a.NotPanic(func() { ctx.NewResult(40000) })
	a.Panic(func() { ctx.NewResult(50000) })
}

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
			4000: "400",
			4001: "401",
		})
	})

	app.Mux().GetFunc("/render", func(w http.ResponseWriter, r *http.Request) {
		ctx := New(w, r, app)
		rslt := ctx.NewResult(4000)
		// 不能使用 SetDetail，顺序未定，可能导致测试失败
		rslt.Add("field1", "message1")
		rslt.Add("field2", "message2")
		rslt.Render()
	})

	app.Mux().GetFunc("/exit", func(w http.ResponseWriter, r *http.Request) {
		ctx := New(w, r, app)
		rslt := ctx.NewResult(4001)
		// 不能使用 SetDetail，顺序未定，可能导致测试失败
		rslt.Add("field1", "message1")
		rslt.Add("field2", "message2")
		rslt.Exit()
	})

	app.Mux().GetFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		ctx := New(w, r, app)
		ctx.NewResult(100).Render()
	})

	srv := rest.NewServer(t, app.Mux(), nil)

	// render 的正常流程测试
	srv.NewRequest(http.MethodGet, "/render").
		Header("Accept", "application/json").
		Do().
		Status(400).
		StringBody(`{"message":"400","code":4000,"detail":[{"field":"field1","message":"message1"},{"field":"field2","message":"message2"}]}`)

	// result.Code 不存在的情况
	srv.NewRequest(http.MethodGet, "/error").
		Header("Accept", "application/json").
		Do().
		Status(http.StatusInternalServerError)

	// result.Exit() 测试
	srv.NewRequest(http.MethodGet, "/exit").
		Header("Accept", "application/json").
		Do().
		Status(400).
		StringBody(`{"message":"401","code":4001,"detail":[{"field":"field1","message":"message1"},{"field":"field2","message":"message2"}]}`)
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
