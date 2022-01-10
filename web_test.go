// SPDX-License-Identifier: MIT

package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert/v2"

	"github.com/issue9/web/serialization/text"
	"github.com/issue9/web/serialization/text/testobject"
)

func TestCreated(t *testing.T) {
	a := assert.New(t, false)
	s, err := NewServer("test", "1.0", nil)
	a.NotError(s.Mimetypes().Add(text.Marshal, text.Unmarshal, text.Mimetype))
	a.NotError(err).NotNil(s)

	w := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodPost, "/path", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", text.Mimetype)
	r.Header.Set("content-type", text.Mimetype)
	resp := Created(&testobject.TextObject{Name: "test", Age: 123}, "")
	ctx := s.NewContext(w, r)
	a.NotError(ctx.Marshal(resp.Status(), resp.Body(), resp.Headers()))
	a.Equal(w.Code, http.StatusCreated).
		Equal(w.Body.String(), `test,123`)

	w.Body.Reset()
	r, err = http.NewRequest(http.MethodPost, "/path", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", text.Mimetype)
	r.Header.Set("content-type", text.Mimetype)
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
	r, err := http.NewRequest(http.MethodPost, "/path", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", text.Mimetype)
	r.Header.Set("content-type", text.Mimetype)
	resp := NotImplemented()
	ctx := s.NewContext(w, r)
	a.NotError(ctx.Marshal(resp.Status(), resp.Body(), resp.Headers()))
	a.Equal(w.Code, http.StatusNotImplemented)

	// Retry-After
	w = httptest.NewRecorder()
	r, err = http.NewRequest(http.MethodPost, "/path", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", text.Mimetype)
	r.Header.Set("content-type", text.Mimetype)
	resp = RetryAfter(http.StatusServiceUnavailable, 120)
	ctx = s.NewContext(w, r)
	a.NotError(ctx.Marshal(resp.Status(), resp.Body(), resp.Headers()))
	a.Equal(w.Code, http.StatusServiceUnavailable).
		Empty(w.Body.String()).
		Equal(w.Header().Get("Retry-After"), "120")

	// Retry-After
	now := time.Now()
	w = httptest.NewRecorder()
	r, err = http.NewRequest(http.MethodPost, "/path", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", text.Mimetype)
	r.Header.Set("content-type", text.Mimetype)
	resp = RetryAt(http.StatusMovedPermanently, now)
	ctx = s.NewContext(w, r)
	a.NotError(ctx.Marshal(resp.Status(), resp.Body(), resp.Headers()))
	a.Equal(w.Code, http.StatusMovedPermanently).
		Empty(w.Body.String()).
		Contains(w.Header().Get("Retry-After"), "GMT")
}
