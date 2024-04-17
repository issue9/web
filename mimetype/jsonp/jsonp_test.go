// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package jsonp

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/web"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

func TestJSONP(t *testing.T) {
	a := assert.New(t, false)
	s, err := server.New("test", "1.0.0", &server.Options{
		Codec:      web.NewCodec().AddMimetype(Mimetype, Marshal, Unmarshal, ""),
		HTTPServer: &http.Server{Addr: ":8080"},
	})
	a.NotError(err).NotNil(s)
	Install(s, "callback", web.ProblemBadRequest)

	s.Routers().New("def", nil).Get("/jsonp", func(ctx *web.Context) web.Responser {
		return web.OK("jsonp")
	})

	defer servertest.Run(a, s)()
	defer s.Close(0)

	servertest.Get(a, "http://localhost:8080/jsonp").Header("accept", Mimetype).Do(nil).
		Status(http.StatusBadRequest).
		BodyEmpty()

	servertest.Get(a, "http://localhost:8080/jsonp?callback=cb").Header("accept", Mimetype).Do(nil).
		StringBody(`cb("jsonp")`)

	servertest.Get(a, "http://localhost:8080/jsonp?cb=cb").Header("accept", Mimetype).Do(nil).
		Status(http.StatusBadRequest).
		BodyEmpty()
}
