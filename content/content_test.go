// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package content

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/web/config"
)

var defaultEnvelopeConf = &config.Envelope{
	State:  config.EnvelopeStateDisable,
	Key:    "",
	Status: http.StatusOK,
}

var envelopeEnableConf = &config.Envelope{
	State:  config.EnvelopeStateEnable,
	Key:    "envelope",
	Status: http.StatusOK,
}

func TestNew(t *testing.T) {
	a := assert.New(t)

	c, err := New("json", envelopeEnableConf)
	a.NotError(err).NotNil(c)
	w := httptest.NewRecorder()
	a.NotNil(w)
	r, err := http.NewRequest("GET", "/index.php?envelope=true", nil)
	r.Header.Set("Accept", jsonContentType)
	a.NotError(err).NotNil(r)

	c.Render(w, r, http.StatusOK, nil, nil)
	a.Equal(w.Body.String(), `{"status":200,"headers":[{"Content-Type":"application/json;charset=utf-8"}]}`)
}
