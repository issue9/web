// SPDX-License-Identifier: MIT

//go:build development

package mode

import (
	"github.com/issue9/web/internal/dev"
	"github.com/issue9/web/server"
)

const defaultMode = Development

func filename(f string) string { dev.Filename(f) }

func debugRouter(r *server.Router, path, id string) {
	dev.DebugRouter(r, path, id)
}
