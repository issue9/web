// SPDX-License-Identifier: MIT

package context

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/message/catalog"
)

func TestContext_Sprintf(t *testing.T) {
	message.SetString(language.MustParse("cmn-hans"), "test", "测试")
	message.SetString(language.MustParse("cmn-hant"), "test", "測試")

	cat := catalog.NewBuilder()
	cat.SetString(language.MustParse("cmn-hans"), "test", "测试1")
	cat.SetString(language.MustParse("cmn-hant"), "test", "測試1")

	a := assert.New(t)

	srv := newServer(a)
	srv.Get("/sprintf", func(ctx *Context) {
		ctx.Render(http.StatusOK, ctx.Sprintf("test"), nil)
	})
	srv.Get("/change", func(ctx *Context) {
		ctx.Server().Catalog = cat
	})
	srv.Get("/fprintf", func(ctx *Context) {
		ctx.Fprintf(ctx.Response, "test")
	})

	s := rest.NewServer(t, srv.Handler(), nil)
	defer s.Close()

	s.Get("/sprintf").
		Header("accept-language", "cmn-hant").
		Header("accept", "application/json").
		Do().
		StringBody(`"測試"`)

	s.Get("/sprintf").
		Header("accept-language", "cmn-hans").
		Header("accept", "application/json").
		Do().
		StringBody(`"测试"`)

	// 切换 catalog
	s.Get("/change").Do().Status(http.StatusOK)

	s.Get("/fprintf").
		Header("accept-language", "cmn-hant").
		Header("accept", "application/json").
		Do().
		StringBody("測試1")

	s.Get("/fprintf").
		Header("accept-language", "cmn-hans").
		Header("accept", "application/json").
		Do().
		StringBody("测试1")
}

func TestContext_Now(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := srv.NewContext(w, r)

	now := ctx.Now()
	a.Equal(now.Location(), srv.Location)
	a.Equal(now.Location(), srv.Now().Location())
}
