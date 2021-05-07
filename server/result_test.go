// SPDX-License-Identifier: MIT

package server

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/web/result"
)

func TestContext_NewResult(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)
	srv.SetErrorHandle(func(w io.Writer, status int) {
		_, err := w.Write([]byte("error-handler"))
		a.NotError(err)
	}, 400) // 此处用于检测是否影响 result.Render() 的输出
	srv.AddResultMessage(400, 40000, "lang") // lang 有翻译
	w := httptest.NewRecorder()

	// 能正常翻译错误信息
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("accept-language", language.SimplifiedChinese.String())
	r.Header.Set("accept", "application/json")
	ctx := srv.NewContext(w, r)
	rslt := ctx.NewResult(40000)
	rslt.Render()
	a.Equal(w.Body.String(), `{"message":"hans","code":40000}`)

	// 未指定 accept-language，采用默认的 und
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("accept", "application/json")
	ctx = srv.NewContext(w, r)
	rslt = ctx.NewResult(40000)
	rslt.Render()
	a.Equal(w.Body.String(), `{"message":"und","code":40000}`)

	// 不存在的本地化信息，采用默认的 und
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("accept-language", "en-US")
	r.Header.Set("accept", "application/json")
	ctx = srv.NewContext(w, r)
	rslt = ctx.NewResult(40000)
	rslt.Render()
	a.Equal(w.Body.String(), `{"message":"und","code":40000}`)

	// 不存在
	a.Panic(func() { ctx.NewResult(400) })
	a.Panic(func() { ctx.NewResult(50000) })
}

func TestContext_NewResultWithFields(t *testing.T) {
	a := assert.New(t)

	r := httptest.NewRequest(http.MethodGet, "/path", bytes.NewBufferString("123"))
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx := newServer(a).NewContext(w, r)
	ctx.server.AddResultMessage(http.StatusBadRequest, 40010, "40010")
	ctx.server.AddResultMessage(http.StatusBadRequest, 40011, "40011")

	rslt := ctx.NewResultWithFields(40010, result.Fields{
		"k1": []string{"v1", "v2"},
	})
	a.True(rslt.HasFields())

	rslt.Render()
	a.Equal(w.Body.String(), `{"message":"40010","code":40010,"fields":[{"name":"k1","message":["v1","v2"]}]}`)
}

func TestServer_ResultMessages(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)
	a.NotNil(srv)

	a.NotPanic(func() {
		srv.AddResultMessage(400, 40010, "lang")
	})

	lmsgs := srv.ResultMessages(message.NewPrinter(language.Und, message.Catalog(srv.catalog)))
	a.Equal(lmsgs[40010], "und")

	lmsgs = srv.ResultMessages(message.NewPrinter(language.SimplifiedChinese, message.Catalog(srv.catalog)))
	a.Equal(lmsgs[40010], "hans")

	lmsgs = srv.ResultMessages(message.NewPrinter(language.TraditionalChinese, message.Catalog(srv.catalog)))
	a.Equal(lmsgs[40010], "hant")

	lmsgs = srv.ResultMessages(message.NewPrinter(language.English, message.Catalog(srv.catalog)))
	a.Equal(lmsgs[40010], "und")
}
