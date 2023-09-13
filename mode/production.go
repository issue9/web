// SPDX-License-Identifier: MIT

//go:build !development

package mode

import "github.com/issue9/web"

const defaultMode = Production

func filename(f string) string { return f }

func debugRouter(r *web.Router, path, id string) {}
