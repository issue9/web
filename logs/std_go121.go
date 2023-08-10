// SPDX-License-Identifier: MIT

//go:build go1.21

package logs

import (
	"log/slog"

	"github.com/issue9/logs/v5"
)

func setStdDefault(l *logs.Logs) {
	slog.SetDefault(l.SLog())
}
