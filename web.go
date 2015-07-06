// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// 一个模块化的微形web框架，依赖于issue9下的其它包。
//  web.Init("~/config")
//  // 其它处理工作...
//  web.Run(errhandler)
//
// 通过web.Init()指定一个配置文件的目录，web包本身有两个配置文件存在于该目录下：
//  web.json 基本的配置内容，包括数据库等信息
//  logs.xml 日志配置信息。
package web

import (
	"net/http"
	"os"

	"github.com/issue9/context"
	"github.com/issue9/logs"
	"github.com/issue9/mux"
)

// 当前库的版本
const Version = "0.1.0.150628"

var serveMux = mux.NewServeMux()

// 初始化web包的内容。
// 若dir目录并不真实存在或其它问题，则会直接panic
func Init(dir string) {
	initConfigDir(dir)

	// 初始化日志系统
	err := logs.InitFromXMLFile(ConfigFile("logs.xml"))
	if err != nil {
		panic(err)
	}

	// 确保在其它需要使用到cfg变量的函数之前调用。
	loadConfig(ConfigFile("web.json"))

	initSession()
	initDB()
	initStatic()
}

func initStatic() {
	group, err := NewModule("static")
	if err != nil {
		panic(err)
	}

	for url, dir := range cfg.Static {
		group.Get(url, http.StripPrefix(url, http.FileServer(http.Dir(dir))))
	}
}

// 开始监听。
// errorHandler 为错误处理函数。
func Run(errHandler mux.RecoverFunc) {
	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if len(cfg.ServerName) > 0 {
			w.Header().Add("Server", cfg.ServerName) // 添加serverName
		}

		serveMux.ServeHTTP(w, req)
		context.Free(req) // 清除context的内容

		// 清除缓存的sessions
		sessionsMu.Lock()
		delete(sessions, req)
		sessionsMu.Unlock()
	})

	if cfg.Https {
		http.ListenAndServeTLS(cfg.Port, cfg.CertFile, cfg.KeyFile, mux.NewRecovery(h, errHandler))
	} else {
		http.ListenAndServe(cfg.Port, mux.NewRecovery(h, errHandler))
	}
}

// 退出程序，退出之前会自动输出所有的日志内容。
func Exit(code int) {
	Flush()
	os.Exit(code)
}
