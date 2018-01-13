// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"time"

	"github.com/issue9/mux"
	"golang.org/x/text/encoding"

	"github.com/issue9/web/app"
	"github.com/issue9/web/context"
	"github.com/issue9/web/modules"
	"github.com/issue9/web/result"
)

var defaultApp *app.App

// Options 传递给 web.Init() 的参数
type Options struct {
	// 配置文件所在的目录。
	ConfigDir string

	// 所有需要支持的字符集
	//
	// 内置 utf-8 字符集支持，不需要再指定
	Charset map[string]encoding.Encoding

	// 指定从客户端数据转换到结构体的函数。
	//
	// 已内置对 text/plain 的支持。
	Marshals map[string]context.Marshal

	// 指定将结构体转换成字符串的函数。
	//
	// 已内置对 text/plain 的支持。
	Unmarshals map[string]context.Unmarshal

	// 所有的错误代码与错误信息的对照表。
	Messages map[int]string
}

// Init 初始化框架的基本内容
func Init(opt *Options) error {
	for name, c := range opt.Charset {
		if err := context.AddCharset(name, c); err != nil {
			return err
		}
	}

	for name, m := range opt.Marshals {
		if err := context.AddMarshal(name, m); err != nil {
			return err
		}
	}

	for name, m := range opt.Unmarshals {
		if err := context.AddUnmarshal(name, m); err != nil {
			return err
		}
	}

	if err := result.NewMessages(opt.Messages); err != nil {
		return err
	}

	app, err := app.New(opt.ConfigDir)
	if err != nil {
		return err
	}

	defaultApp = app
	return nil
}

// Run 运行路由，执行监听程序，具体说明可参考 App.Run()。
func Run(build app.BuildHandler) error {
	return defaultApp.Run(build)
}

// Shutdown 关闭所有服务，具体说明可参考 App.Shutdown()
func Shutdown(timeout time.Duration) error {
	return defaultApp.Shutdown(timeout)
}

// File 获取配置目录下的文件，具体说明可参考 App.File()。
func File(path string) string {
	return defaultApp.File(path)
}

// Router 获取操作路由的接口，为一个 mux.Prefix 实例，具体接口说明可参考 issue9/mux 包。
func Router() *mux.Prefix {
	return defaultApp.Router()
}

// URL 构建一条基于 Config.Root 的完整 URL
func URL(path string) string {
	return defaultApp.URL(path)
}

// Module 注册一个新的模块，具体说明可参考 App.Mux()。
func Module(name string, init modules.InitFunc, deps ...string) {
	defaultApp.Module(name, init, deps...)
}

// NewContext 根据当前配置，生成 context.Context 对象，若是出错则返回 nil
func NewContext(w http.ResponseWriter, r *http.Request) *context.Context {
	return defaultApp.NewContext(w, r)
}

// NewResult 生成一个 *result.Result 对象
func NewResult(code int, fields map[string]string) *result.Result {
	return result.New(code, fields)
}
