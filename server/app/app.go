// SPDX-License-Identifier: MIT

// Package app 提供了简便的方式管理 [web.Server]
package app

import (
	"sync"
	"time"

	"github.com/issue9/web"
)

// App 简单的 [web.Server] 管理
type App struct {
	// 构建新服务的方法
	//
	// 每次重启服务时，都将由此方法生成一个新的服务。
	// 只有在返回成功的新实例时，才会替换旧实例，否则旧实例将一直运行。
	NewServer func() (web.Server, error)
	srv       web.Server
	srvLock   sync.RWMutex

	// 每次关闭服务操作的等待时间
	ShutdownTimeout time.Duration

	restart     bool
	restartLock sync.Mutex
}

func (app *App) init() (err error) {
	if app.NewServer == nil {
		panic("app.NewServer 不能为空")
	}

	app.srv, err = app.NewServer()
	return err
}

// Exec 运行服务
func (app *App) Exec() error {
	if err := app.init(); err != nil {
		return err
	}

RESTART:
	app.restart = false
	err := app.srv.Serve()
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

	// 先拿到旧服务，以便在新服务初始化失败时能正确输出日志。
	app.srvLock.RLock()
	old := app.srv
	app.srvLock.RUnlock()

	srv, err := app.NewServer()
	if err != nil {
		old.Logs().ERROR().Error(err)
		return
	}
	app.srvLock.Lock()
	app.srv = srv
	app.srvLock.Unlock()

	old.Close(app.ShutdownTimeout) // 新服务声明成功，尝试关闭旧服务。
}
