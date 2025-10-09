// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package app

// Restart 根据信号 c 重启 a
//
// 可以结合其它方法一起使用，比如和 [fsnotify] 一起使用：
//
//	watcher := fsnotify.NewWatcher(...)
//	Restart(watcher.Event, s)
//
// 也可参考 [SignalHUP]。
//
// [fsnotify]: https://pkg.go.dev/github.com/fsnotify/fsnotify
func Restart[T any](c chan T, a App) {
	go func() {
		for range c {
			a.Restart()
		}
	}()
}
