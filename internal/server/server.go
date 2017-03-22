// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

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

	// mux 的外层封装。
	handler http.Handler

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

// Init 初始化。
//
// h 表示需要执行的路由处理函数，传递 nil 时，会自动以 Server.Mux() 代替。
// 可以通过以下方式，将一些 http.Handler 实例附加到 Server.Mux() 之上：
//  s.Init(handlers.Host(s.Mux(), "www.caixw.io")
func (s *Server) Init(h http.Handler) {
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
	s.handler = s.buildHandler(h)
}

// Restart 重启服务。
func (s *Server) Restart(timeout time.Duration) error {
	if err := s.Shutdown(timeout); err != nil {
		return err
	}

	return s.Run()
}

// Run 运行路由，执行监听程序。
func (s *Server) Run() error {
	if s.conf.HTTPS {
		switch s.conf.HTTPState {
		case httpStateListen:
			logs.Infof("开始监听[%v]端口", httpPort)
			go s.getServer(httpPort, s.handler).ListenAndServe()
		case httpStateRedirect:
			logs.Infof("开始监听[%v]端口，并跳转至[%v]", httpPort, httpsPort)
			go s.httpRedirectListenAndServe()
			// 空值或是 disable 均为默认处理方式，即不作为。
		}

		logs.Infof("开始监听[%v]端口", s.conf.Port)
		return s.getServer(s.conf.Port, s.handler).ListenAndServeTLS(s.conf.CertFile, s.conf.KeyFile)
	}

	logs.Infof("开始监听[%v]端口", s.conf.Port)
	return s.getServer(s.conf.Port, s.handler).ListenAndServe()
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
func (s *Server) httpRedirectListenAndServe() error {
	srv := s.getServer(httpPort, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL
		url.Host = r.Host + s.conf.Port
		url.Scheme = "HTTPS"

		urlStr := url.String()
		http.Redirect(w, r, urlStr, http.StatusMovedPermanently)
	}))

	return srv.ListenAndServe()
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
