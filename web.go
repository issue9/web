// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"os"

	"github.com/issue9/context"
	"github.com/issue9/logs"
	"github.com/issue9/mux"
)

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

	// 加载配置文件的内容
	loadConfig(ConfigFile("web.json"))

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
		delete(sessions, req)
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
