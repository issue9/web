// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

//go:build !js

package app

import (
	"os"
	"os/signal"
	"syscall"
)

// SignalHUP 让 a 根据 [HUP] 信号重启服务
//
//	app := &App{...}
//	SignalHUP(app)
//
// [HUP]: https://en.wikipedia.org/wiki/SIGHUP
func SignalHUP(a App) {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGHUP)
	Restart(sc, a)
}
