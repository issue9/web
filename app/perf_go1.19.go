// SPDX-License-Identifier: MIT

//go:build go1.19

package app

import "runtime/debug"

func initMemoryLimit(size int64) { debug.SetMemoryLimit(size) }
