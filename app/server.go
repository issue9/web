// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package app

import (
	"net/http"
	"strings"

	"github.com/issue9/logs"
	"github.com/issue9/middleware/compress"
)

// 初始化路由项。
func (app *App) initRoutes(h http.Handler) http.Handler {
	// 静态文件路由，在其它路由构建之前调用
	for url, dir := range app.config.Static {
		pattern := url + "{path}"
		app.mux.Get(pattern, http.StripPrefix(url, compress.New(http.FileServer(http.Dir(dir)), logs.ERROR())))
	}

	if h == nil {
		h = app.mux
	}

	// 构建其它路由
	return app.buildHandler(h)
}

// 运行路由，执行监听程序。
//
// h 表示需要执行的路由处理函数，传递 nil 时，会自动以 server.Mux() 代替。
// 可以通过以下方式，将一些 http.Handler 实例附加到 server.Mux() 之上：
//  app.run(host.New(app.Mux(), "www.caixw.io")
func (app *App) run(h http.Handler) error {
	h = app.initRoutes(h)

	if app.config.HTTPS {
		switch app.config.HTTPState {
		case httpStateListen:
			go func() {
				logs.Infof("开始监听[%v]端口", httpPort)
				logs.Error(app.getServer(httpPort, h).ListenAndServe())
			}()
		case httpStateRedirect:
			go func() {
				logs.Infof("开始监听[%v]端口，并跳转至[%v]", httpPort, httpsPort)
				logs.Error(app.httpRedirectServer().ListenAndServe())
			}()
			// 空值或是 disable 均为默认处理方式，即不作为。
		}

		logs.Infof("开始监听[%v]端口", app.config.Port)
		return app.getServer(app.config.Port, h).ListenAndServeTLS(app.config.CertFile, app.config.KeyFile)
	}

	logs.Infof("开始监听[%v]端口", app.config.Port)
	return app.getServer(app.config.Port, h).ListenAndServe()
}

// 构建一个从 HTTP 跳转到 HTTPS 的路由服务。
func (app *App) httpRedirectServer() *http.Server {
	return app.getServer(httpPort, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 构建跳转链接
		url := r.URL
		url.Scheme = "HTTPS"
		url.Host = strings.Split(r.Host, ":")[0] + app.config.Port

		urlStr := url.String()
		http.Redirect(w, r, urlStr, http.StatusMovedPermanently)
	}))
}

// 获取 http.Server 实例，相对于 http 的默认实现，指定了 ErrorLog 字段。
func (app *App) getServer(port string, h http.Handler) *http.Server {
	srv := &http.Server{
		Addr:         port,
		Handler:      h,
		ErrorLog:     logs.ERROR(),
		ReadTimeout:  app.config.ReadTimeout,
		WriteTimeout: app.config.WriteTimeout,
	}

	app.servers = append(app.servers, srv)

	return srv
}
