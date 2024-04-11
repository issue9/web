// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package app

import (
	"os"
	"os/signal"
	"syscall"
)

// ServerApp 定义了服务类型的 APP 的接口
//
// [App] 和 [CLI] 均实现了此接口。
type ServerApp interface {
	// RestartServer 重启 APP
	//
	// 中止旧的 [web.Server]，再启动一个新的 [web.Server] 对象。
	//
	// NOTE: 如果执行过程中出错，应该尽量阻止旧对象被中止，保证最大限度地可用状态。
	RestartServer()
}

// SignalHUP 让 s 根据 [HUP] 信号重启服务
//
//	app := &App{...}
//	SignalHUP(app)
//
// [HUP]: https://en.wikipedia.org/wiki/SIGHUP
func SignalHUP(s ServerApp) {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGHUP)
	Restart(sc, s)
}

// Restart 根据信号 c 重启 s
//
// 可以结合其它方法一起使用，比如和 fsnotify 一起使用：
//
//	watcher := fsnotify.NewWatcher(...)
//	Restart(watcher.Event, s)
//
// 也可参考 [SignalHUP]。
func Restart[T any](c chan T, s ServerApp) {
	go func() {
		for range c {
			s.RestartServer()
		}
	}()
}
