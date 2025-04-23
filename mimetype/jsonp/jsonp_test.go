// SPDX-FileCopyrightText: 2018-2025 caixw
//
// SPDX-License-Identifier: MIT

package jsonp

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/mux/v9/header"

	"github.com/issue9/web"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

func TestJSONP(t *testing.T) {
	a := assert.New(t, false)
	s, err := server.NewHTTP("test", "1.0.0", &server.Options{
		Codec:      web.NewCodec().AddMimetype(Mimetype, Marshal, Unmarshal, "", true, true),
		HTTPServer: &http.Server{Addr: ":8080"},
	})
	a.NotError(err).NotNil(s)
	Install(s, "callback", web.ProblemBadRequest)

	s.Routers().New("def", nil).Get("/jsonp", func(ctx *web.Context) web.Responser {
		return web.OK("jsonp")
	})

	defer servertest.Run(a, s)()
	defer s.Close(0)

	servertest.Get(a, "http://localhost:8080/jsonp").Header(header.Accept, Mimetype).Do(nil).
		Status(http.StatusBadRequest).
		BodyEmpty()

	servertest.Get(a, "http://localhost:8080/jsonp?callback=cb").Header(header.Accept, Mimetype).Do(nil).
		StringBody(`cb("jsonp")`)

	servertest.Get(a, "http://localhost:8080/jsonp?cb=cb").Header(header.Accept, Mimetype).Do(nil).
		Status(http.StatusBadRequest).
		BodyEmpty()
}
