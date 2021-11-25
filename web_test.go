// SPDX-License-Identifier: MIT

package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v2"

	"github.com/issue9/web/serialization/text"
	"github.com/issue9/web/serialization/text/testobject"
)

func TestCreated(t *testing.T) {
	a := assert.New(t, false)
	w := httptest.NewRecorder()
	s, err := NewServer("test", "1.0", nil)
	a.NotError(s.Mimetypes().Add(text.Marshal, text.Unmarshal, text.Mimetype))
	a.NotError(err).NotNil(s)

	r := httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set("Accept", text.Mimetype)
	r.Header.Set("content-type", text.Mimetype)
	resp := Created(&testobject.TextObject{Name: "test", Age: 123}, "")
	ctx := s.NewContext(w, r)
	a.NotError(ctx.Marshal(resp.Status(), resp.Body(), resp.Headers()))
	a.Equal(w.Code, http.StatusCreated).
		Equal(w.Body.String(), `test,123`)

	w.Body.Reset()
	r = httptest.NewRequest(http.MethodPost, "/path", nil)
	r.Header.Set("Accept", text.Mimetype)
	r.Header.Set("content-type", text.Mimetype)
	resp = Created(&testobject.TextObject{Name: "test", Age: 123}, "/test")
	ctx = s.NewContext(w, r)
	a.NotError(ctx.Marshal(resp.Status(), resp.Body(), resp.Headers()))
	a.Equal(w.Code, http.StatusCreated).
		Equal(w.Body.String(), `test,123`).
		Equal(w.Header().Get("Location"), "/test")
}
