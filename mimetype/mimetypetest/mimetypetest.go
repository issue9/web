// SPDX-FileCopyrightText: 2025-2026 caixw
//
// SPDX-License-Identifier: MIT

// Package mimetypetest 提供了用于测试 mimetype 的函数
package mimetypetest

import (
	"bytes"
	"net/http"
	"net/http/httptest"

	"github.com/issue9/assert/v6"
	"github.com/issue9/mux/v10/types"
	"golang.org/x/text/language"

	"github.com/issue9/web"
	"github.com/issue9/web/server"
	"github.com/issue9/web/startuptest"
)

var inst = &web.Problem{
	Type:     "https://example.com/probs/test",
	Title:    "test",
	Status:   200,
	Detail:   "test",
	Instance: "instance",
}

func Test(a *assert.Assertion, m web.MarshalFunc, u web.UnmarshalFunc) {
	ts := startuptest.NewTestServerStartup(a)
	s, err := server.NewHTTP("test", "1.0", &server.Options{
		Startup:  ts,
		Language: language.English,
	})
	a.NotError(err).NotNil(s)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	ctx := s.NewContext(w, r, types.NewRoute())

	data, err := m(ctx, inst)
	a.NotError(err).NotNil(data)

	inst2 := &web.Problem{}
	a.NotError(u(bytes.NewBuffer(data), inst2))
	a.Equal(inst, inst2)
}
