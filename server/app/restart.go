// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

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

// Restart 根据信号 c 重启 a
//
// 可以结合其它方法一起使用，比如和 fsnotify 一起使用：
//
//	watcher := fsnotify.NewWatcher(...)
//	Restart(watcher.Event, s)
//
// 也可参考 [SignalHUP]。
func Restart[T any](c chan T, a App) {
	go func() {
		for range c {
			a.Restart()
		}
	}()
}
