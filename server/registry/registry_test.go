// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package registry_test

import (
	"net/http"
	"os"

	"github.com/issue9/assert/v4"
	"github.com/issue9/logs/v7"

	"github.com/issue9/web"
	"github.com/issue9/web/server"
)

func newTestServer(a *assert.Assertion) web.Server {
	srv, err := server.NewHTTP("test", "1.0.0", &server.Options{
		Logs:       logs.New(logs.NewTermHandler(os.Stderr, nil), logs.WithLevels(logs.AllLevels()...), logs.WithLocation(true), logs.WithCreated(logs.NanoLayout)),
		HTTPServer: &http.Server{Addr: ":8080"},
	})

	a.NotError(err).NotNil(srv)
	return srv
}
