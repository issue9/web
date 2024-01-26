// SPDX-License-Identifier: MIT

// Package app 提供了简便的方式管理 [web.Server]
package app

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/issue9/web"
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

// App 简单的 [web.Server] 管理
type App struct {
	// 构建新服务的方法
	//
	// 每次重启服务时，都将由此方法生成一个新的服务。
	NewServer func() (web.Server, error)
	srv       web.Server

	// 重启之前需要做的操作
	//
	// 可以为空。如果返回值不为 nil，将中止当前操作，但不影响旧服务。
	Before func() error

	// 每次关闭服务操作的等待时间
	ShutdownTimeout time.Duration

	restart     bool
	restartLock sync.Mutex
}

// Exec 运行服务
func (app *App) Exec() (err error) {
RESTART:
	app.srv, err = app.NewServer()
	if err != nil {
		return web.NewStackError(err)
	}

	app.restart = false
	err = app.srv.Serve()
	if app.restart { // 等待 Serve 过程中，如果调用 RestartServer，会将 app.restart 设置为 true。
		goto RESTART
	}
	return err
}

// RestartServer 触发重启服务
//
// 该方法将关闭现有的服务，并发送运行新服务的指令，不会等待新服务启动完成。
func (app *App) RestartServer() {
	app.restartLock.Lock()
	defer app.restartLock.Unlock()

	app.restart = true

	if err := app.Before(); err != nil {
		app.srv.Logs().ERROR().Error(err)
		return
	}

	app.srv.Close(app.ShutdownTimeout)
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
