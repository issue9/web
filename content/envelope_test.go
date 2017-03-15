// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package content

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestNewEnvelope(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()

	e := newEnvelope(http.StatusOK, w.Header(), nil)
	a.NotNil(e)
	a.Equal(e.Status, http.StatusOK).Equal(len(e.Headers), 0).Nil(e.Response)

	// 检测 Header, w.Header 会将 Header 中的首字母大写
	w.Header().Set("key1", "val1")
	e = newEnvelope(http.StatusTeapot, w.Header(), nil)
	a.NotNil(e)
	a.Equal(e.Status, http.StatusTeapot).Equal(len(e.Headers), 1).Nil(e.Response)

	// 检测 Response
	w = httptest.NewRecorder()
	e = newEnvelope(http.StatusOK, w.Header(), "123")
	a.NotNil(e)
	a.Equal(e.Status, http.StatusOK).Equal(len(e.Headers), 0).Equal(e.Response, "123")
}
