// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

//go:build !development

package comptime

const defaultMode = Production

func filename(f string) string { return f }
