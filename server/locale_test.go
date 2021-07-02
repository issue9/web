// SPDX-License-Identifier: MIT

package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
	"github.com/issue9/mux/v5/group"
	"golang.org/x/text/language"
)

func TestContext_Sprintf(t *testing.T) {
	a := assert.New(t)

	srv := newServer(a)
	a.NotError(srv.Content().CatalogBuilder().SetString(language.MustParse("cmn-hans"), "test", "测试"))
	a.NotError(srv.Content().CatalogBuilder().SetString(language.MustParse("cmn-hant"), "test", "測試"))
	router, err := srv.NewRouter("default", "http://localhost:8081/root", group.MatcherFunc(group.Any))
	a.NotError(err).NotNil(router)

	router.Get("/sprintf", func(ctx *Context) {
		ctx.Render(http.StatusOK, ctx.Sprintf("test"), nil)
	})

	s := rest.NewServer(t, srv.groups, nil)
	defer s.Close()

	s.Get("/root/sprintf").
		Header("accept-language", "cmn-hant").
		Header("accept", "application/json").
		Do().
		StringBody(`"測試"`)

	s.Get("/root/sprintf").
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
	ctx := srv.NewContext(w, r)

	now := ctx.Now()
	a.Equal(now.Location(), srv.Location())
	a.Equal(now.Location(), srv.Now().Location())
}
