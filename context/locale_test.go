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
)

func TestContext_Sprint(t *testing.T) {
	message.SetString(language.MustParse("cmn-hans"), "test", "测试")
	message.SetString(language.MustParse("cmn-hant"), "test", "測試")

	a := assert.New(t)

	srv := newServer(a)
	srv.Get("/test", func(ctx *Context) {
		ctx.Render(http.StatusOK, ctx.Sprintf("test"), nil)
	})
	s := rest.NewServer(t, srv.Handler(), nil)
	defer s.Close()

	s.Get("/test").
		Header("accept-language", "cmn-hant").
		Header("accept", "application/json").
		Do().
		StringBody(`"測試"`)

	s.Get("/test").
		Header("accept-language", "cmn-hans").
		Header("accept", "application/json").
		Do().
		StringBody(`"测试"`)
}

func TestContext_Now(t *testing.T) {
	a := assert.New(t)
	srv := newServer(a)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := srv.newContext(w, r)

	now := ctx.Now()
	a.Equal(now.Location(), srv.Location)
	a.Equal(now.Location(), srv.Now().Location())
}
