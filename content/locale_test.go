// SPDX-License-Identifier: MIT

package content

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert"
	"golang.org/x/text/language"

	"github.com/issue9/web/content/text"
	"github.com/issue9/web/serialization"
)

func TestContent_catalog(t *testing.T) {
	a := assert.New(t)

	c := New(DefaultBuilder, time.Local, serialization.NewFiles(10), language.SimplifiedChinese)
	a.NotNil(c)

	err := c.Locale().Builder().SetString(language.SimplifiedChinese, "test", "测试")
	a.NotError(err)
	err = c.Locale().Builder().SetString(language.Und, "test", "und")
	a.NotError(err)

	p := c.newLocalePrinter(language.SimplifiedChinese)
	a.Equal(p.Sprintf("test"), "测试")
	p = c.newLocalePrinter(language.Japanese)
	a.Equal(p.Sprintf("test"), "und")
}

func TestContext_LocalePrinter(t *testing.T) {
	a := assert.New(t)
	c := New(DefaultBuilder, time.Local, serialization.NewFiles(10), language.SimplifiedChinese)
	a.NotNil(c)
	a.NotError(c.Mimetypes().Add(text.Marshal, text.Unmarshal, text.Mimetype))
	a.NotError(c.Mimetypes().Add(text.Marshal, text.Unmarshal, DefaultMimetype))

	b := c.Locale().Builder()
	a.NotError(b.SetString(language.MustParse("cmn-hans"), "test", "测试"))
	a.NotError(b.SetString(language.MustParse("cmn-hant"), "test", "測試"))

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
