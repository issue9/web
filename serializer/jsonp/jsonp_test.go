// SPDX-License-Identifier: MIT

package jsonp

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web"

	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

func TestJSONP(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewTester(a, &server.Options{
		Mimetypes: []*server.Mimetype{
			{Type: Mimetype, Marshal: Marshal, Unmarshal: Unmarshal, ProblemType: ""},
		},
		HTTPServer: &http.Server{Addr: ":8080"},
	})
	Install("callback", s.Server())

	s.Router().Get("/jsonp", func(ctx *server.Context) server.Responser {
		return web.OK("jsonp")
	})

	s.GoServe()

	s.Get("/jsonp").Header("accept", Mimetype).Do(nil).
		StringBody(`"jsonp"`)

	s.Get("/jsonp?callback=cb").Header("accept", Mimetype).Do(nil).
		StringBody(`cb("jsonp")`)

	s.Get("/jsonp?cb=cb").Header("accept", Mimetype).Do(nil).
		StringBody(`"jsonp"`)

	s.Close(0)
}
