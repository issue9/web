// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

// Package mimetypetest 提供了用于测试 mimetype 的函数
package mimetypetest

import (
	"bytes"
	"net/http"

	"github.com/issue9/assert/v4"
	"golang.org/x/text/language"

	"github.com/issue9/web"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

var inst = &web.Problem{
	Type:     "https://example.com/probs/test",
	Title:    "test",
	Status:   200,
	Detail:   "test",
	Instance: "instance",
}

func Test(a *assert.Assertion, m web.MarshalFunc, u web.UnmarshalFunc) {
	s, err := server.NewHTTP("test", "1.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Language:   language.English,
	})
	a.NotError(err).NotNil(s)

	s.Routers().New("main", nil).Get("/path", func(ctx *web.Context) web.Responser {
		data, err := m(ctx, inst)
		a.NotError(err).NotNil(data)

		inst2 := &web.Problem{}
		a.NotError(u(bytes.NewBuffer(data), inst2))
		a.Equal(inst, inst2)

		return web.OK(nil)
	})

	defer servertest.Run(a, s)()
	defer s.Close(0)

	servertest.Get(a, "http://localhost:8080/path").
		Do(nil).
		Status(http.StatusOK)
}
