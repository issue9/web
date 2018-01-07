// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package context

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestRenderStatus(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()

	RenderStatus(w, http.StatusOK)
	a.Equal(w.Code, http.StatusOK).
		Equal(w.Header().Get("Content-Type"), "text/plain; charset=utf-8")

	w = httptest.NewRecorder()
	RenderStatus(w, http.StatusInternalServerError)
	a.Equal(w.Code, http.StatusInternalServerError).
		Equal(w.Header().Get("Content-Type"), "text/plain; charset=utf-8")
}
