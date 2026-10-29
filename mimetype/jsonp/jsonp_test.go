// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package jsonp

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v6"
	"github.com/issue9/mux/v10/header"

	"github.com/issue9/web"
	"github.com/issue9/web/server"
	"github.com/issue9/web/startuptest"
)

func TestJSONP(t *testing.T) {
	a := assert.New(t, false)
	ts := startuptest.NewTestServerStartup(a)
	s, err := server.NewHTTP("test", "1.0.0", &server.Options{
		Codec:   web.NewCodec().AddMimetype(Mimetype, Marshal, Unmarshal, "", true, true),
		Startup: ts,
	})
	a.NotError(err).NotNil(s)
	Install(s, "callback", web.ProblemBadRequest)

	s.Routers().New("def", nil).Get("/jsonp", func(ctx *web.Context) web.Responser {
		return web.OK("jsonp")
	})

	ts.Get("/jsonp").Header(header.Accept, Mimetype).Do().
		Status(http.StatusBadRequest).
		BodyEmpty()

	ts.Get("/jsonp?callback=cb").Header(header.Accept, Mimetype).Do().
		StringBody(`cb("jsonp")`)

	ts.Get("/jsonp?cb=cb").Header(header.Accept, Mimetype).Do().
		Status(http.StatusBadRequest).
		BodyEmpty()
}
