// SPDX-License-Identifier: MIT

package filters

import (
	"net/http"

	"github.com/issue9/web/server"
)

func ContentTypeFilter(ct ...string) server.Filter {
	return func(next server.HandlerFunc) server.HandlerFunc {
		return ContentType(next, ct...)
	}
}

func ContentType(next server.HandlerFunc, ct ...string) server.HandlerFunc {
	return func(ctx *server.Context) server.Responser {
		for _, c := range ct {
			if c == ctx.OutputMimetypeName {
				return next(ctx)
			}
		}
		return server.Status(http.StatusNotAcceptable)
	}
}
