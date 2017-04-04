// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func testParams(a *assert.Assertion, pattern, path string, h http.Handler) {
	app, err := NewApp("./testdata")
	a.NotError(err).NotNil(app)
	app.Mux().Get(pattern, h)

	srv := httptest.NewServer(app.Mux())
	defer srv.Close()

	resp, err := http.Get(srv.URL + path)
	a.NotError(err).NotNil(resp)
}

func TestParams_Int(t *testing.T) {
	a := assert.New(t)

	buildHandler := func(vals map[string]int) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := app.NewContext(w, r)
			for key, val := range vals {
				a.Equal(val, ctx.MustInt(key))
			}
		})
	}

	testParams(a, "/test/{id:\\d+}/{id2:\\d+}", "/test/123/456", http.HandlerFunc(f1))
}
