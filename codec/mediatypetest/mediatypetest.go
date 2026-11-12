// SPDX-FileCopyrightText: 2025-2026 caixw
//
// SPDX-License-Identifier: MIT

// Package mediatypetest 提供了用于测试 [web.MediaType] 的函数
package mediatypetest

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/issue9/assert/v6"
	"github.com/issue9/cache/drivers/memory"
	"github.com/issue9/config"
	"github.com/issue9/logs/v7"
	"github.com/issue9/mux/v10/header"
	"github.com/issue9/mux/v10/types"
	"golang.org/x/text/language"

	"github.com/issue9/web"
	"github.com/issue9/web/startuptest"
)

var inst = &web.Problem{
	Type:     "https://example.com/probs/test",
	Title:    "test",
	Status:   200,
	Detail:   "test",
	Instance: "instance",
}

func Test(a *assert.Assertion, mt *web.MediaType) {
	ts := startuptest.NewTestServerStartup(a)
	conf := config.Dir(nil, "./")
	log := logs.New(logs.NewNopHandler())
	unique := func() string { return "" }
	cc := memory.New()
	s := web.NewServer("test", "1.0.0", conf, ts, time.Local, language.SimplifiedChinese, log, unique, cc, web.NewCodec(), header.XRequestID)
	a.NotNil(s)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	ctx := s.NewContext(w, r, types.NewRoute())

	data, err := mt.Marshal(ctx, inst)
	a.NotError(err).NotNil(data)

	inst2 := &web.Problem{}
	a.NotError(mt.Unmarshal(bytes.NewBuffer(data), inst2))
	a.Equal(inst, inst2)
}
