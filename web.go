// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"time"

	"github.com/issue9/web/context"
	"github.com/issue9/web/result"
)

var defaultApp *App

// BuildHandler 将一个 http.Handler 封装成另一个 http.Handler
//
// 一个中间件的接口定义，传递给 New() 函数，可以给全部的路由项添加一个中间件。
type BuildHandler func(http.Handler) http.Handler

// Init 初始化框架的基本内容
func Init() (err error) {
	defaultApp, err = NewApp()
	return err
}

// IsDebug 是否处在调试模式
func IsDebug() bool {
	return defaultApp.IsDebug()
}

// Run 运行路由，执行监听程序。
func Run() error {
	return defaultApp.Run()
}

// Shutdown 关闭所有服务。
func Shutdown(timeout time.Duration) error {
	return defaultApp.Shutdown(timeout)
}

// File 获取配置目录下的文件。
func File(path string) string {
	return defaultApp.File(path)
}

// URL 构建一条完整 URL
func URL(path string) string {
	return defaultApp.URL(path)
}

// AddModule 注册一个模块
func AddModule(m *Module) *App {
	return defaultApp.AddModule(m)
}

// NewContext 根据当前配置，生成 context.Context 对象，若是出错则返回 nil
func NewContext(w http.ResponseWriter, r *http.Request) *context.Context {
	return defaultApp.NewContext(w, r)
}

// NewResult 生成一个 *result.Result 对象
func NewResult(code int) *result.Result {
	return result.New(code, nil)
}

// NewResultWithDetail 声明一个带详细内容的 result.Result 对象
func NewResultWithDetail(code int, fields map[string]string) *result.Result {
	return result.New(code, fields)
}
