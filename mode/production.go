// SPDX-License-Identifier: MIT

//go:build !development

package mode

import "github.com/issue9/web/server"

const defaultMode = Production

func filename(f string) string { return f }

func debugRouter(r *server.Router, path, id string) {}
