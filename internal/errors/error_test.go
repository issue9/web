// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package errors

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/web/encoding"
)

func TestTraceStack(t *testing.T) {
	a := assert.New(t)

	str := TraceStack(1, "message", 12)
	a.True(strings.HasPrefix(str, "message12"))
	a.True(strings.Contains(str, "error_test.go")) // 肯定包含当前文件名
}

func TestRenderStatus(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()

	RenderStatus(w, http.StatusOK)
	a.Equal(w.Code, http.StatusOK).
		Equal(w.Header().Get("Content-Type"), encoding.BuildContentType(encoding.DefaultMimeType, encoding.DefaultCharset))

	w = httptest.NewRecorder()
	RenderStatus(w, http.StatusInternalServerError)
	a.Equal(w.Code, http.StatusInternalServerError).
		Equal(w.Header().Get("Content-Type"), encoding.BuildContentType(encoding.DefaultMimeType, encoding.DefaultCharset))
}
