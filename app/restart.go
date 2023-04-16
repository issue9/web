// SPDX-License-Identifier: MIT

package app

import (
	"os"
	"os/signal"
	"syscall"
)

// RestartServer 实现对 [server.Server] 的重启操作
//
// 利用该接口可以实现很多操作，比如根据信号重启服务，或是监听配置文件而重启等。
type RestartServer interface {
	// Restart 重启服务
	//
	// 中止旧的 [server.Server]，再启动一个新的 [server.Server] 对象
	Restart()
}

// SignalHUP 根据 [HUP] 信号重启服务
//
// [HUP]: https://en.wikipedia.org/wiki/SIGHUP
func SignalHUP(cmd RestartServer) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGHUP)

	go func() {
		for range signalChannel {
			cmd.Restart()
		}
	}()
}
