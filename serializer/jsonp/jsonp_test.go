// SPDX-License-Identifier: MIT

package jsonp

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web"

	"github.com/issue9/web/servertest"
)

func TestJSONP(t *testing.T) {
	a := assert.New(t, false)
	s, err := web.NewServer("test", "1.0.0", &web.Options{
		Mimetypes: []*web.Mimetype{
			{Type: Mimetype, MarshalBuilder: BuildMarshal, Unmarshal: Unmarshal, ProblemType: ""},
		},
		HTTPServer: &http.Server{Addr: ":8080"},
	})
	a.NotError(err).NotNil(s)
	Install("callback", s)

	s.NewRouter("def", nil).Get("/jsonp", func(ctx *web.Context) web.Responser {
		return web.OK("jsonp")
	})

	defer servertest.Run(a, s)()
	defer s.Close(0)

	servertest.Get(a, "http://localhost:8080/jsonp").Header("accept", Mimetype).Do(nil).
		StringBody(`"jsonp"`)

	servertest.Get(a, "http://localhost:8080/jsonp?callback=cb").Header("accept", Mimetype).Do(nil).
		StringBody(`cb("jsonp")`)

	servertest.Get(a, "http://localhost:8080/jsonp?cb=cb").Header("accept", Mimetype).Do(nil).
		StringBody(`"jsonp"`)
}
