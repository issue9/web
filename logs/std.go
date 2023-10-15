// SPDX-License-Identifier: MIT

//go:build !go1.21

package logs

import (
	"log"

	"github.com/issue9/logs/v6"
	"github.com/issue9/logs/v6/writers"
)

func setStdDefault(l *logs.Logs) {
	log.SetOutput(writers.WriteFunc(func(b []byte) (int, error) {
		l.INFO().String(string(b))
		return len(b), nil
	}))
}
