// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

//go:build !development

package comptime

import "github.com/issue9/web"

const defaultMode = Production

func filename(f string) string { return f }

func debugRouter(r *web.Router, path, id string) {}
