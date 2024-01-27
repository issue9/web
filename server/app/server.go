// SPDX-License-Identifier: MIT

package app

import (
	"os"
	"os/signal"
	"syscall"
)

// ServerApp 实现对 [web.Server] 的管理
type ServerApp interface {
	// RestartServer 重启服务
	//
	// 中止旧的 [web.Server]，再启动一个新的 [web.Server] 对象。
	//
	// 如果执行过程中出错，应该尽量阻止旧对象被中止，保证最大限度地可用状态。
	RestartServer()
}

// SignalHUP 让 s 根据 [HUP] 信号重启服务
//
// [HUP]: https://en.wikipedia.org/wiki/SIGHUP
func SignalHUP(s ServerApp) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGHUP)

	go func() {
		for range signalChannel {
			s.RestartServer()
		}
	}()
}
