// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

//go:build development

package comptime

import "github.com/issue9/web/internal/dev"

const defaultMode = Development

func filename(f string) string { dev.Filename(f) }
