// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"errors"
	"net/http"

	"github.com/issue9/middleware"
	"github.com/issue9/web/context"
	"github.com/issue9/web/internal/app"
	"github.com/issue9/web/module"
	"github.com/issue9/web/result"
)

var defaultApp *app.App

// Init 初始化整个应用环境
//
// configDir 表示配置文件的目录；
func Init(configDir string) (err error) {
	if defaultApp != nil {
		return errors.New("不能重复调用 Init")
	}

	defaultApp, err = app.New(configDir)
	return
}

// SetMiddleware 设置一个全局的中间件，多次设置，只有最后一次会启作用。
func SetMiddleware(m middleware.Middleware) {
	defaultApp.SetMiddleware(m)
}

// IsDebug 是否处在调试模式
func IsDebug() bool {
	return defaultApp.Debug()
}

// Handler 将当前实例当作一个 http.Handler 返回。一般用于测试。
// 比如在 httptest.NewServer 中使用。
func Handler() (http.Handler, error) {
	return defaultApp.Handler()
}

// Run 运行路由，执行监听程序。
func Run() error {
	return defaultApp.Serve()
}

// Close 关闭服务。
//
// 无论配置文件如果设置，此函数都是直接关闭服务，不会等待。
func Close() error {
	return defaultApp.Close()
}

// Shutdown 关闭所有服务。
//
// 根据配置文件中的配置项，决定当前是直接关闭还是延时之后关闭。
func Shutdown() error {
	return defaultApp.Shutdown()
}

// File 获取配置目录下的文件。
func File(path string) string {
	return defaultApp.File(path)
}

// URL 构建一条完整 URL
func URL(path string) string {
	return defaultApp.URL(path)
}

// Modules 当前系统使用的所有模块信息
func Modules() []*module.Module {
	return defaultApp.Modules()
}

// NewModule 注册一个模块
func NewModule(name, desc string, deps ...string) (*module.Module, error) {
	return defaultApp.NewModule(name, desc, deps...)
}

// NewContext 根据当前配置，生成 context.Context 对象，若是出错则返回 nil
func NewContext(w http.ResponseWriter, r *http.Request) *context.Context {
	return defaultApp.NewContext(w, r)
}

// NewResult 生成一个 *result.Result 对象
func NewResult(code int) *result.Result {
	return result.New(code, nil)
}
