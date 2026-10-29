// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package registry_test

import (
	"os"

	"github.com/issue9/assert/v6"
	"github.com/issue9/logs/v7"

	"github.com/issue9/web"
	"github.com/issue9/web/server"
	"github.com/issue9/web/startuptest"
)

func newTestServer(a *assert.Assertion) (*web.Server, startuptest.Startup) {
	startup := startuptest.NewHTTPServerStartup(a, false)
	srv, err := server.NewHTTP("test", "1.0.0", &server.Options{
		Logs:    logs.New(logs.NewTermHandler(os.Stderr, nil), logs.WithLevels(logs.AllLevels()...), logs.WithLocation(true), logs.WithCreated(logs.NanoLayout)),
		Startup: startup,
	})

	a.NotError(err).NotNil(srv)
	return srv, startup
}
