// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package server 与路由有关的服务管理。
//
// NOTE: 测试内容涉及到 80 端口，部分系统可能需要 sudo go test 才能正确执行。
package server

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/issue9/handlers"
	"github.com/issue9/logs"
	"github.com/issue9/mux"
)

// Server 服务器控制
type Server struct {
	conf *Config
	mux  *mux.ServeMux

	// 除了 mux 所依赖的 http.Server 实例之外，
	// 还有诸如 80 端口跳转等产生的 http.Server 实例。
	// 记录这些 server，方便重启等操作。
	servers []*http.Server
}

// New 声明一个新的 Server 实例
func New(conf *Config) (*Server, error) {
	if conf == nil {
		return nil, errors.New("参数 conf 不能为空")
	}

	if err := conf.Sanitize(); err != nil {
		return nil, err
	}

	return &Server{
		mux:     mux.NewServeMux(!conf.Options),
		servers: make([]*http.Server, 0, 5),
		conf:    conf,
	}, nil
}

// Mux 返回默认的 *mux.ServeMux 实例
func (s *Server) Mux() *mux.ServeMux {
	return s.mux
}

// 初始化路由项。
func (s *Server) initRoutes(h http.Handler) http.Handler {
	// 静态文件路由，在其它路由构建之前调用
	for url, dir := range s.conf.Static {
		if !strings.HasSuffix(url, "/") {
			url += "/"
		}
		s.mux.Get(url, http.StripPrefix(url, handlers.Compress(http.FileServer(http.Dir(dir)))))
	}

	if h == nil {
		h = s.mux
	}

	// 构建其它路由
	return s.buildHandler(h)
}

// Run 运行路由，执行监听程序。
//
// h 表示需要执行的路由处理函数，传递 nil 时，会自动以 Server.Mux() 代替。
// 可以通过以下方式，将一些 http.Handler 实例附加到 Server.Mux() 之上：
//  s.Run(handlers.Host(s.Mux(), "www.caixw.io")
func (s *Server) Run(h http.Handler) error {
	h = s.initRoutes(h)

	if s.conf.HTTPS {
		switch s.conf.HTTPState {
		case httpStateListen:
			go func() {
				logs.Infof("开始监听[%v]端口", httpPort)
				logs.Error(s.getServer(httpPort, h).ListenAndServe())
			}()
		case httpStateRedirect:
			go func() {
				logs.Infof("开始监听[%v]端口，并跳转至[%v]", httpPort, httpsPort)
				logs.Error(s.httpRedirectServer().ListenAndServe())
			}()
			// 空值或是 disable 均为默认处理方式，即不作为。
		}

		logs.Infof("开始监听[%v]端口", s.conf.Port)
		return s.getServer(s.conf.Port, h).ListenAndServeTLS(s.conf.CertFile, s.conf.KeyFile)
	}

	logs.Infof("开始监听[%v]端口", s.conf.Port)
	return s.getServer(s.conf.Port, h).ListenAndServe()
}

// Shutdown 关闭服务。
//
// timeout 若超过该时间，服务还未自动停止的，则会强制停止。
// 若 timeout<=0，则会立即停止服务；
// 若 timeout>0 时，则会等待处理完毕或是该时间耗尽才停止服务。
func (s *Server) Shutdown(timeout time.Duration) error {
	if timeout <= 0 {
		for _, srv := range s.servers {
			if err := srv.Close(); err != nil {
				return err
			}
		}
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		for _, srv := range s.servers {
			if err := srv.Shutdown(ctx); err != nil {
				return err
			}
		}
	}

	s.servers = s.servers[:0]
	return nil
}

// 构建一个从 HTTP 跳转到 HTTPS 的路由服务。
func (s *Server) httpRedirectServer() *http.Server {
	return s.getServer(httpPort, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 构建跳转链接
		url := r.URL
		url.Scheme = "HTTPS"
		url.Host = strings.Split(r.Host, ":")[0] + s.conf.Port

		urlStr := url.String()
		http.Redirect(w, r, urlStr, http.StatusMovedPermanently)
	}))
}

// 获取 http.Server 实例，相对于 http 的默认实现，指定了 ErrorLog 字段。
func (s *Server) getServer(port string, h http.Handler) *http.Server {
	srv := &http.Server{
		Addr:         port,
		Handler:      h,
		ErrorLog:     logs.ERROR(),
		ReadTimeout:  s.conf.ReadTimeout,
		WriteTimeout: s.conf.WriteTimeout,
	}

	s.servers = append(s.servers, srv)

	return srv
}
