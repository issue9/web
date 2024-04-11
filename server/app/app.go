// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package app 提供了简便的方式管理 [web.Server]
package app

import (
	"sync"
	"time"

	"github.com/issue9/web"
)

// App [ServerApp] 的简单实现
type App struct {
	// 构建新服务的方法
	//
	// 每次重启服务时，都将由此方法生成一个新的服务。
	// 只有在返回成功的新实例时，才会替换旧实例，否则旧实例将一直运行。
	NewServer func() (web.Server, error)

	// 每次关闭服务操作的等待时间
	ShutdownTimeout time.Duration

	srv     web.Server
	srvLock sync.RWMutex
	exit    chan struct{} // 保证只有在前一个 srv 退出之后才会让 RestartServer 退出

	restart           bool
	restartServerLock sync.Mutex
}

func (app *App) getServer() web.Server {
	app.srvLock.RLock()
	s := app.srv
	app.srvLock.RUnlock()
	return s
}

func (app *App) setServer(s web.Server) {
	app.srvLock.Lock()
	app.srv = s
	app.srvLock.Unlock()
}

// Exec 运行服务
func (app *App) Exec() (err error) {
	if app.NewServer == nil {
		panic("App.NewServer 不能为空")
	}

	if app.srv, err = app.NewServer(); err != nil {
		return
	}
	app.exit = make(chan struct{}, 1)

RESTART:
	app.restart = false
	err = app.getServer().Serve()
	app.exit <- struct{}{} // 在 Serve 返回之后再通知 RestartServer 可以返回了。
	if app.restart {       // 等待 Serve 过程中，如果调用 RestartServer，会将 app.restart 设置为 true。
		goto RESTART
	}
	return err
}

// RestartServer 触发重启服务
//
// 该方法将关闭现有的服务，并发送运行新服务的指令，不会等待新服务启动完成。
func (app *App) RestartServer() {
	app.restartServerLock.Lock()
	defer app.restartServerLock.Unlock()

	app.restart = true

	old := app.getServer()

	srv, err := app.NewServer()
	if err != nil {
		old.Logs().ERROR().Error(err)
		return
	}
	app.setServer(srv)

	old.Close(app.ShutdownTimeout) // 新服务声明成功，尝试关闭旧服务。
	<-app.exit                     // 等待 server.Serve 退出
}
