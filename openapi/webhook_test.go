// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/web"
)

func TestDocument_AddWebHook(t *testing.T) {
	a := assert.New(t, false)
	ss := newServer(a)
	d := New(ss, web.Phrase("desc"))

	d.AddWebhook("hook1", http.MethodGet, &Operation{})
	a.Length(d.webHooks, 1)

	a.PanicString(func() {
		d.AddWebhook("hook1", http.MethodGet, &Operation{})
	}, "已经存在 hook1:GET 的 webhook")
	a.Length(d.webHooks, 1)

	d.AddWebhook("hook1", http.MethodPost, &Operation{})
	a.Length(d.webHooks, 1)
}
