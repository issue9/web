// SPDX-License-Identifier: MIT

package registry_test

import (
	"net/http"
	"os"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web"
	"github.com/issue9/web/server"
)

func newTestServer(a *assert.Assertion) web.Server {
	srv, err := server.New("test", "1.0.0", &server.Options{
		Logs: &server.Logs{
			Handler:  server.NewTermHandler(os.Stderr, nil),
			Created:  server.NanoLayout,
			Location: true,
			Levels:   server.AllLevels(),
		},
		HTTPServer: &http.Server{Addr: ":8080"},
	})

	a.NotError(err).NotNil(srv)
	return srv
}
