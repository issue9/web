// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package result

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/middleware/recovery"
	"golang.org/x/text/language"
	xmessage "golang.org/x/text/message"

	"github.com/issue9/web/context"
	"github.com/issue9/web/encoding/form"
	"github.com/issue9/web/internal/errors"
)

var _ form.Marshaler = &Result{}

func TestResult_Add_HasDetail(t *testing.T) {
	a := assert.New(t)

	code := 400 * 1000
	a.NotError(NewMessage(code, "400"))
	r := &Result{Code: code}
	a.False(r.HasDetail())

	r.Add("field", "message")
	r.Add("field", "message")
	a.True(r.HasDetail())
	a.Equal(len(r.Detail), 2)

	cleanMessage()
}

func TestResult_SetDetail(t *testing.T) {
	a := assert.New(t)

	code := 400 * 1000
	a.NotError(NewMessage(code, "400"))
	r := &Result{Code: code}
	a.False(r.HasDetail())

	r.SetDetail(map[string]string{"field1": "message1", "field2": "message2"})
	r.SetDetail(map[string]string{"field1": "message1", "field2": "message2"})
	a.True(r.HasDetail())
	a.Equal(len(r.Detail), 2)

	cleanMessage()
}

func TestResult_Render(t *testing.T) {
	a := assert.New(t)
	code := http.StatusForbidden * 1000
	a.NotError(NewMessage(code, "400"))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := &context.Context{
			OutputMimeType:     json.Marshal,
			OutputMimeTypeName: "application/json",
			Request:            r,
			Response:           w,
			LocalePrinter:      xmessage.NewPrinter(language.Und),
		}
		rslt := &Result{Code: code}
		rslt.SetDetail(map[string]string{"field1": "message1", "field2": "message2"})
		rslt.Render(ctx)
	}))

	r, err := http.NewRequest(http.MethodGet, srv.URL+"/path", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(r)
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusForbidden)

	cleanMessage()
}

func TestResult_Render_error(t *testing.T) {
	a := assert.New(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := &context.Context{
			OutputMimeType:     json.Marshal,
			OutputMimeTypeName: "application/json",
			Request:            r,
			Response:           w,
			LocalePrinter:      xmessage.NewPrinter(language.Und),
		}
		rslt := &Result{Code: 100}
		rslt.Render(ctx)
	}))

	r, err := http.NewRequest(http.MethodGet, srv.URL+"/path", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(r)
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusInternalServerError)

	cleanMessage()
}

func TestResult_Exit(t *testing.T) {
	a := assert.New(t)
	code := http.StatusForbidden * 1000
	a.NotError(NewMessage(code, "400"))

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := &context.Context{
			OutputMimeType:     json.Marshal,
			OutputMimeTypeName: "application/json",
			Request:            r,
			Response:           w,
			LocalePrinter:      xmessage.NewPrinter(language.Und),
		}
		rslt := &Result{Code: code}
		rslt.SetDetail(map[string]string{"field1": "message1", "field2": "message2"})
		rslt.Exit(ctx)
	})

	srv := httptest.NewServer(recovery.New(h, errors.Recovery(false)))
	r, err := http.NewRequest(http.MethodGet, srv.URL+"/path", nil)
	a.NotError(err).NotNil(r)
	r.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(r)
	a.NotError(err).NotNil(resp)
	a.Equal(resp.StatusCode, http.StatusForbidden)

	cleanMessage()
}
