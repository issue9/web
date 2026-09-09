// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

// Package app 提供了简便的方式管理 [web.Server] 的运行
package app

import (
	"slices"
	"sync"
	"time"

	"github.com/issue9/localeutil"
	"github.com/kardianos/service"

	"github.com/issue9/web"
)

// App [web.Server] 的管理接口
type App interface {
	// Exec 运行当前程序
	Exec() error

	// Restart 重启
	//
	// 中止旧的 [web.Server]，再启动一个新的 [web.Server] 对象。
	//
	// NOTE: 如果执行过程中出错，应该尽量阻止旧对象被中止，保证最大限度地可用状态。
	Restart()
}

type app struct {
	// 构建新服务的方法
	//
	// 每次重启服务时，都将由此方法生成一个新的服务。
	// 只有在返回成功的新实例时，才会替换旧实例，否则旧实例将一直运行。
	newServer func() (web.Server, error)

	shutdownTimeout time.Duration

	srv     web.Server
	srvLock sync.RWMutex
	exit    chan struct{} // 保证只有在前一个 srv 退出之后才会让 RestartServer 退出

	restart           bool
	restartServerLock sync.Mutex
}

// New 声明一个简要的 [App] 实现
//
// shutdown 每次关闭服务操作的等待时间；
// newServer 构建新服务的方法；
func New(shutdown time.Duration, newServer func() (web.Server, error)) App {
	return newApp(shutdown, newServer)
}

func newApp(shutdown time.Duration, newServer func() (web.Server, error)) *app {
	if newServer == nil {
		panic("newServer 不能为空")
	}

	return &app{
		shutdownTimeout: shutdown,
		newServer:       newServer,
		exit:            make(chan struct{}, 1),
	}
}

func (app *app) getServer() web.Server {
	app.srvLock.RLock()
	s := app.srv
	app.srvLock.RUnlock()
	return s
}

func (app *app) setServer(s web.Server) {
	app.srvLock.Lock()
	app.srv = s
	app.srvLock.Unlock()
}

// Exec 运行服务
func (app *app) Exec() (err error) {
	if app.srv, err = app.newServer(); err != nil {
		return err
	}

RESTART:
	app.restart = false
	err = app.getServer().Serve()
	app.exit <- struct{}{} // 在 Serve 返回之后再通知 RestartServer 可以返回了。
	if app.restart {       // 等待 Serve 过程中，如果调用 RestartServer，会将 app.restart 设置为 true。
		goto RESTART
	}
	return err
}

// Restart 触发重启服务
//
// 该方法将关闭现有的服务，并发送运行新服务的指令，不会等待新服务启动完成。
func (app *app) Restart() {
	app.restartServerLock.Lock()
	defer app.restartServerLock.Unlock()

	app.restart = true

	old := app.getServer()

	srv, err := app.newServer()
	if err != nil {
		old.Logs().ERROR().Error(err)
		return
	}
	app.setServer(srv)

	old.Close(app.shutdownTimeout) // 新服务声明成功，尝试关闭旧服务。
	<-app.exit                     // 等待 server.Serve 退出
}

type DaemonConfig struct {
	DisplayName  web.LocaleStringer
	Description  web.LocaleStringer
	UserName     string
	Arguments    []string
	Executable   string   // 服务的可执行文件路径。
	Dependencies []string // 服务依赖的其他服务。

	// 以下字段不支持 windows

	WorkingDirectory string // 服务的初始工作目录。
	ChRoot           string
	Option           map[string]any // 系统特定选项。
	EnvVars          map[string]string
}

func (d *DaemonConfig) toServiceConfig(name string, p *localeutil.Printer) *service.Config {
	return &service.Config{
		Name:             name,
		DisplayName:      d.DisplayName.LocaleString(p),
		Description:      d.Description.LocaleString(p),
		UserName:         d.UserName,
		Arguments:        d.Arguments,
		Executable:       d.Executable,
		Dependencies:     d.Dependencies,
		WorkingDirectory: d.WorkingDirectory,
		ChRoot:           d.ChRoot,
		Option:           d.Option,
		EnvVars:          d.EnvVars,
	}
}

// 执行守护进程功能并返回当前的状态
//
// action 可以是 [service.ControlAction] 和 'status' 中的任意元素；
func (app *app) runDaemon(action string, conf *service.Config) (service.Status, error) {
	if !isDaemonControl(action) {
		return service.StatusUnknown, web.NewLocaleError("invalid daemon control action: %s", action)
	}

	d, err := service.New(app, conf)
	if err != nil {
		return service.StatusUnknown, err
	}

	if action != "status" {
		if err := service.Control(d, action); err != nil {
			return service.StatusUnknown, err
		}
	}
	if action == "uninstall" { // 已卸载，无法获取状态。
		return service.StatusUnknown, nil
	}
	return d.Status()
}

// 只使用了 [service.Control] 的功能，不需要实现 [service.Interface] 的具体功能。
func (app *app) Start(s service.Service) error { return nil }
func (app *app) Stop(s service.Service) error  { return nil }

func statusString(s service.Status) string {
	switch s {
	case service.StatusRunning:
		return "running"
	case service.StatusStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

func isDaemonControl(c string) bool {
	return c == "status" || slices.Contains(service.ControlAction[:], c)
}
