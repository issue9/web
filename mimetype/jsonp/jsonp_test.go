// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package jsonp

import (
	"net/http"
	"sync"
	"testing"

	"github.com/issue9/assert/v6"
	"github.com/issue9/mux/v10/header"

	"github.com/issue9/web"
	"github.com/issue9/web/problems"
	"github.com/issue9/web/server"
	"github.com/issue9/web/startuptest"
)

func TestJSONP(t *testing.T) {
	once = &sync.Once{} // 防止 go test 在 -count 大于 1 的情况下失败

	a := assert.New(t, false)
	ts := startuptest.NewTestServerStartup(a)
	s, err := server.NewHTTP("test", "1.0.0", &server.Options{
		Codec:   web.NewCodec().AddMediaType(web.NewMediaType(Mimetype, "", Marshal, Unmarshal), true, true),
		Startup: ts,
	})
	a.NotError(err).NotNil(s)
	Install(s, "callback", problems.BadRequest)

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
