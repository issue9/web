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

var defaultConf = &Config{
	ContentType:    "json",
	EnvelopeState:  EnvelopeStateDisable,
	EnvelopeKey:    "",
	EnvelopeStatus: http.StatusOK,
}

var envelopeEnableConf = &Config{
	ContentType:    "json",
	EnvelopeState:  EnvelopeStateEnable,
	EnvelopeKey:    "envelope",
	EnvelopeStatus: http.StatusOK,
}

func TestNew(t *testing.T) {
	a := assert.New(t)

	c, err := New(envelopeEnableConf)
	a.NotError(err).NotNil(c)
	w := httptest.NewRecorder()
	a.NotNil(w)
	r, err := http.NewRequest("GET", "/index.php?envelope=true", nil)
	r.Header.Set("Accept", jsonContentType)
	a.NotError(err).NotNil(r)

	c.Render(w, r, http.StatusOK, nil, nil)
	a.Equal(w.Body.String(), `{"status":200,"headers":[{"Content-Type":"application/json;charset=utf-8"}]}`)
}
