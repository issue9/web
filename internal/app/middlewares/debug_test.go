// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package middlewares

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/rest"
)

func TestDebug(t *testing.T) {
	srv := rest.NewServer(t, debug(h202), nil)

	// 命中 /debug/pprof/cmdline
	srv.NewRequest(http.MethodGet, "/debug/pprof/").
		Do().
		Status(http.StatusOK)

	srv.NewRequest(http.MethodGet, "/debug/pprof/cmdline").
		Do().
		Status(http.StatusOK)

	srv.NewRequest(http.MethodGet, "/debug/pprof/trace").
		Do().
		Status(http.StatusOK)

	srv.NewRequest(http.MethodGet, "/debug/pprof/symbol").
		Do().
		Status(http.StatusOK)

	// /debug/vars
	srv.NewRequest(http.MethodGet, "/debug/vars").
		Do().
		Status(http.StatusOK)

	// 命中 h202
	srv.NewRequest(http.MethodGet, "/debug/").
		Do().
		Status(http.StatusAccepted)
}
