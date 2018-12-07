// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package result

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
	"github.com/issue9/middleware/recovery"
	"golang.org/x/text/language"
	xmessage "golang.org/x/text/message"

	"github.com/issue9/web/app"
	"github.com/issue9/web/context"
	"github.com/issue9/web/mimetype"
	"github.com/issue9/web/mimetype/form"
)

var (
	_ form.Marshaler = &Result{}
	_ error          = &Result{}
)

// 声明一个 App 实例
func newApp(a *assert.Assertion) *app.App {
	app, err := app.New(&app.Options{
		Dir: "../testdata",

		MimetypeMarshals: map[string]mimetype.MarshalFunc{
			"application/json": json.Marshal,
			"application/xml":  xml.Marshal,
		},

		MimetypeUnmarshals: map[string]mimetype.UnmarshalFunc{
			"application/json": json.Unmarshal,
			"application/xml":  xml.Unmarshal,
		},
	})

	a.NotError(err).NotNil(app)

	return app
}

func TestResult_Add_HasDetail(t *testing.T) {
	a := assert.New(t)

	code := 400 * 1000
	a.NotError(NewMessage(code, "400"))
	r := &Result{Code: code}
	a.False(r.HasDetail())

	r.Add("field", "message")
	r.Add("field", "message")
	a.True(r.HasDetail())
	a.Equal(len(r.Detail), 2)

	cleanMessage()
}

func TestResult_SetDetail(t *testing.T) {
	a := assert.New(t)

	code := 400 * 1000
	a.NotError(NewMessage(code, "400"))
	r := &Result{Code: code}
	a.False(r.HasDetail())

	r.SetDetail(map[string]string{"field1": "message1", "field2": "message2"})
	r.SetDetail(map[string]string{"field1": "message1", "field2": "message2"})
	a.True(r.HasDetail())
	a.Equal(len(r.Detail), 2)

	cleanMessage()
}

func TestResult_Render_Exit(t *testing.T) {
	a := assert.New(t)

	a.NotError(NewMessages(map[int]string{
		http.StatusForbidden * 1000:    "400", // 需要与 resultRenderHandler 中的错误代码值相同
		http.StatusUnauthorized * 1000: "401", // 需要与 resultRenderHandler 中的错误代码值相同
	}))

	resultRenderHandler := func(w http.ResponseWriter, r *http.Request) {
		ctx := &context.Context{
			OutputMimeType:     json.Marshal,
			OutputMimeTypeName: "application/json",
			Request:            r,
			Response:           w,
			LocalePrinter:      xmessage.NewPrinter(language.Und),
		}

		switch r.URL.Path {
		case "/render":
			rslt := &Result{Code: http.StatusForbidden * 1000}
			rslt.SetDetail(map[string]string{"field1": "message1", "field2": "message2"})
			rslt.Render(ctx)
		case "/exit":
			rslt := &Result{Code: http.StatusUnauthorized * 1000}
			rslt.SetDetail(map[string]string{"field1": "message1", "field2": "message2"})
			rslt.Exit(ctx)
		case "/error":
			rslt := &Result{Code: 100}
			rslt.Render(ctx)
		}
	}

	h := http.HandlerFunc(resultRenderHandler)
	srv := rest.NewServer(t, recovery.New(h, recoverFunc), nil)

	// render 的正常流程测试
	srv.NewRequest(http.MethodGet, "/render").
		Do().
		Status(http.StatusForbidden)

	// result.Code 不存在的情况
	srv.NewRequest(http.MethodGet, "/error").
		Do().
		Status(http.StatusInternalServerError) // 等同于 recoverFunc 中的输出报头

	// result.Exit() 测试
	srv.NewRequest(http.MethodGet, "/exit").
		Do().
		Status(http.StatusUnauthorized)

	cleanMessage()
}

// 仅作为 TestResult_Render_Exit 的测试用
func recoverFunc(w http.ResponseWriter, msg interface{}) {
	w.WriteHeader(http.StatusInternalServerError)
}
