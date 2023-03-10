// SPDX-License-Identifier: MIT

// Package app 为构建程序提供相对简便的方法
//
// 提供了以下几种方式构建服务：
//   - [App] 简要的 [server.Server] 管理对象，需要用户指定 [server.Server] 的生成方式；
//   - [NewServerOf] 从配置文件构建 [server.Server] 对象；
//   - [CLIOf] 直接生成一个简单的命令行程序，结合了 [App] 和 [NewServerOf] 两者的功能；
//
// NOTE: 这并不一个必需的包，如果觉得不适用，可以直接采用 [server.New] 初始化服务。
//
// # 配置文件
//
// [NewServerOf] 和 [CLIOf] 都是通过加载配置文件对项目进行初始化。
// 对于配置文件各个字段的定义，可参考源代码，入口在 config.go 文件的 configOf 对象。
// 配置文件中除了固定的字段之外，还提供了泛型变量 User 用于指定用户自定义的额外字段。
//
// # 注册函数
//
// 当前包提供大量的注册函数，以用将某些无法直接采用序列化的内容转换可序列化的。
// 比如通过 [RegisterEncoding] 将 `gzip-default` 等字符串表示成压缩算法，
// 以便在配置文件进行指定。
//
// 所有的注册函数处理逻辑上都相似，碰上同名的会覆盖，否则是添加。
// 且默认情况下都提供了一些可选项，只有在用户需要额外添加自己的内容时才需要调用注册函数。
package app

import (
	"sync"
	"time"

	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/server"
)

// App 简要的服务管理功能
type App struct {
	// 构建新服务的方法
	NewServer func() (*server.Server, error)
	srv       *server.Server

	// 重启之前需要做的操作
	//
	// 如果返回值不为 nil，将退出当前的 Restart 操作，但不影响旧服务。
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
		return errs.NewStackError(err)
	}

	app.restart = false
	err = app.srv.Serve()
	if app.restart { // 等待 Serve 过程中，如果调用 Restart，会将 cmd.restart 设置为 true。
		goto RESTART
	}
	return err
}

// Restart 触发重启服务
//
// 该方法将关闭现有的服务，并发送运行新服务的指令，不会等待新服务启动完成。
func (app *App) Restart() {
	app.restartLock.Lock()
	defer app.restartLock.Unlock()

	app.restart = true

	if err := app.Before(); err != nil {
		app.srv.Logs().ERROR().Error(err)
		return
	}

	app.srv.Close(app.ShutdownTimeout)
}
