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
	"github.com/issue9/web/server/servertest"
)

func TestCreated(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewServer(a, nil)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", nil).
		Header("Accept", text.Mimetype).
		Header("content-type", text.Mimetype).
		Request()
	ctx := s.NewContext(w, r)
	resp := Created(&testobject.TextObject{Name: "test", Age: 123}, "")
	a.NotError(resp.Apply(ctx))
	a.Equal(w.Code, http.StatusCreated).
		Equal(w.Body.String(), `test,123`)

	w = httptest.NewRecorder()
	r = rest.Post(a, "/path", nil).
		Header("Accept", text.Mimetype).
		Header("content-type", text.Mimetype).
		Request()
	resp = Created(&testobject.TextObject{Name: "test", Age: 123}, "/test")
	ctx = s.NewContext(w, r)
	a.NotError(resp.Apply(ctx))
	a.Equal(w.Code, http.StatusCreated).
		Equal(w.Body.String(), `test,123`).
		Equal(w.Header().Get("Location"), "/test")
}

func TestStatus(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewServer(a, nil)

	w := httptest.NewRecorder()
	r := rest.Post(a, "/path", nil).
		Header("Accept", text.Mimetype).
		Header("content-type", text.Mimetype).
		Request()
	resp := NotImplemented()
	ctx := s.NewContext(w, r)
	a.NotError(resp.Apply(ctx))
	a.Equal(w.Code, http.StatusNotImplemented)

	// Retry-After
	w = httptest.NewRecorder()
	r = rest.Post(a, "/path", nil).
		Header("Accept", text.Mimetype).
		Header("content-type", text.Mimetype).
		Request()
	ctx = s.NewContext(w, r)
	resp = RetryAfter(http.StatusServiceUnavailable, 120)
	a.NotError(resp.Apply(ctx))
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
	a.NotError(resp.Apply(ctx))
	a.Equal(w.Code, http.StatusMovedPermanently).
		Empty(w.Body.String()).
		Contains(w.Header().Get("Retry-After"), "GMT")
}

func TestRedirect(t *testing.T) {
	a := assert.New(t, false)

	r := rest.Post(a, "/path", []byte("123")).
		Header("Accept", "application/json").
		Header("Content-Type", "application/json").
		Request()
	w := httptest.NewRecorder()
	ctx := servertest.NewServer(a, nil).NewContext(w, r)
	resp := Redirect(301, "https://example.com")
	a.NotError(resp.Apply(ctx))

	a.Equal(w.Result().StatusCode, 301).
		Equal(w.Header().Get("Location"), "https://example.com")
}
