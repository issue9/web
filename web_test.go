// SPDX-License-Identifier: MIT

package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"

	"github.com/issue9/web/serialization/text"
	"github.com/issue9/web/serialization/text/testobject"
)

func TestCreated(t *testing.T) {
	a := assert.New(t, false)
	s, err := NewServer("test", "1.0", nil)
	a.NotError(s.Mimetypes().Add(text.Marshal, text.Unmarshal, text.Mimetype))
	a.NotError(err).NotNil(s)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", nil).
		Header("Accept", text.Mimetype).
		Header("content-type", text.Mimetype).
		Request()
	ctx := s.NewContext(w, r)
	resp := Created(&testobject.TextObject{Name: "test", Age: 123}, "")
	a.NotError(ctx.Marshal(resp.Status(), resp.Body(), resp.Headers()))
	a.Equal(w.Code, http.StatusCreated).
		Equal(w.Body.String(), `test,123`)

	w.Body.Reset()
	r = rest.Post(a, "/path", nil).
		Header("Accept", text.Mimetype).
		Header("content-type", text.Mimetype).
		Request()
	resp = Created(&testobject.TextObject{Name: "test", Age: 123}, "/test")
	ctx = s.NewContext(w, r)
	a.NotError(ctx.Marshal(resp.Status(), resp.Body(), resp.Headers()))
	a.Equal(w.Code, http.StatusCreated).
		Equal(w.Body.String(), `test,123`).
		Equal(w.Header().Get("Location"), "/test")
}

func TestStatus(t *testing.T) {
	a := assert.New(t, false)
	s, err := NewServer("test", "1.0", nil)
	a.NotError(s.Mimetypes().Add(text.Marshal, text.Unmarshal, text.Mimetype))
	a.NotError(err).NotNil(s)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", nil).
		Header("Accept", text.Mimetype).
		Header("content-type", text.Mimetype).
		Request()
	resp := NotImplemented()
	ctx := s.NewContext(w, r)
	a.NotError(ctx.Marshal(resp.Status(), resp.Body(), resp.Headers()))
	a.Equal(w.Code, http.StatusNotImplemented)

	// Retry-After
	w = httptest.NewRecorder()
	r = rest.Post(a, "/path", nil).
		Header("Accept", text.Mimetype).
		Header("content-type", text.Mimetype).
		Request()
	ctx = s.NewContext(w, r)
	resp = RetryAfter(http.StatusServiceUnavailable, 120)
	a.NotError(ctx.Marshal(resp.Status(), resp.Body(), resp.Headers()))
	a.Equal(w.Code, http.StatusServiceUnavailable).
		Empty(w.Body.String()).
		Equal(w.Header().Get("Retry-After"), "120")

	// Retry-After
	now := time.Now()
	w = httptest.NewRecorder()
	r = rest.Post(a, "/path", nil).
		Header("Accept", text.Mimetype).
		Header("content-type", text.Mimetype).
		Request()
	ctx = s.NewContext(w, r)
	resp = RetryAt(http.StatusMovedPermanently, now)
	a.NotError(ctx.Marshal(resp.Status(), resp.Body(), resp.Headers()))
	a.Equal(w.Code, http.StatusMovedPermanently).
		Empty(w.Body.String()).
		Contains(w.Header().Get("Retry-After"), "GMT")
}
