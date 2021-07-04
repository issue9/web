// SPDX-License-Identifier: MIT

package content

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/web/content/text"
	"golang.org/x/text/language"
)

func TestContent_catalog(t *testing.T) {
	a := assert.New(t)

	c := New(DefaultBuilder)
	a.NotNil(c)
	l := c.LocaleBuilder(language.SimplifiedChinese)
	l.SetString("test", "测试")
	l = c.LocaleBuilder(language.Und)
	l.SetString("test", "und")

	p := c.NewLocalePrinter(language.SimplifiedChinese)
	a.Equal(p.Sprintf("test"), "测试")
	p = c.NewLocalePrinter(language.Japanese)
	a.Equal(p.Sprintf("test"), "und")
}

func TestContext_LocalePrinter(t *testing.T) {
	a := assert.New(t)
	c := New(DefaultBuilder)
	a.NotNil(c)
	a.NotError(c.AddMimetype(text.Mimetype, text.Marshal, text.Unmarshal))
	a.NotError(c.AddMimetype(DefaultMimetype, text.Marshal, text.Unmarshal))

	a.NotError(c.CatalogBuilder().SetString(language.MustParse("cmn-hans"), "test", "测试"))
	a.NotError(c.CatalogBuilder().SetString(language.MustParse("cmn-hant"), "test", "測試"))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("accept-language", "cmn-hant")
	r.Header.Set("accept", text.Mimetype)
	ctx, status := c.NewContext(nil, w, r)
	a.Empty(status).NotNil(ctx)
	a.NotError(ctx.Marshal(http.StatusOK, ctx.Sprintf("test"), nil))
	a.Equal(w.Body.String(), "測試")

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("accept-language", "cmn-hans")
	r.Header.Set("accept", text.Mimetype)
	ctx, status = c.NewContext(nil, w, r)
	a.Empty(status).NotNil(ctx)
	n, err := ctx.Fprintf(ctx.Response, "test")
	a.NotError(err).Equal(n, len("测试"))
	a.Equal(w.Body.String(), "测试")
}
