// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"net/http"
	"net/http/pprof"
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
func New(conf *Config) *Server {
	// NOTE: conf 只能来自 web.go，理论上可以保证
	// 其一直是正确的，所以此处不再做正确的检测。
	return &Server{
		mux:     mux.NewServeMux(!conf.Options),
		servers: make([]*http.Server, 0, 5),
		conf:    conf,
	}
}

// Mux 返回默认的 *mux.ServeMux 实例
func (s *Server) Mux() *mux.ServeMux {
	return s.mux
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

// Init 初始化路由项
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

		logs.Infof("开始监听%v端口", s.conf.Port)
		return s.getServer(s.conf.Port, s.handler).ListenAndServeTLS(s.conf.CertFile, s.conf.KeyFile)
	}

	logs.Infof("开始监听%v端口", s.conf.Port)
	return s.getServer(s.conf.Port, s.handler).ListenAndServe()
}

// 构建一个从 HTTP 跳转到 HTTPS 的路由服务。
func (s *Server) httpRedirectListenAndServe() error {
	srv := s.getServer(httpPort, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 替换原 URL 的端口和 Scheme
		url := r.URL
		url.Host = r.Host + s.conf.Port
		url.Scheme = "HTTPS"

		urlStr := url.String()
		http.Redirect(w, r, urlStr, http.StatusMovedPermanently)
	}))

	return srv.ListenAndServe()
}

func (s *Server) buildHandler(h http.Handler) http.Handler {
	h = s.buildHeader(h)

	h = handlers.Recovery(h, func(w http.ResponseWriter, msg interface{}) {
		logs.Error(msg)
	})

	// NOTE: 在最外层添加调试地址，保证调试内容不会被其它 handler 干扰。
	if len(s.conf.Pprof) > 0 {
		return s.buildPprof(h)
	}

	return h
}

// 根据 Config.Pprof 决定是否包装调试地址，调用前请确认是否已经开启 Pprof 选项
func (s *Server) buildPprof(h http.Handler) http.Handler {
	logs.Debug("web:", "开启了调试功能，地址为：", s.conf.Pprof)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, s.conf.Pprof) {
			h.ServeHTTP(w, r)
			return
		}

		path := r.URL.Path[len(s.conf.Pprof):]
		switch path {
		case "cmdline":
			pprof.Cmdline(w, r)
		case "profile":
			pprof.Profile(w, r)
		case "symbol":
			pprof.Symbol(w, r)
		case "trace":
			pprof.Trace(w, r)
		default:
			pprof.Index(w, r)
		}
	}) // end return http.HandlerFunc
}

func (s *Server) buildHeader(h http.Handler) http.Handler {
	if len(s.conf.Headers) == 0 {
		return h
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range s.conf.Headers {
			w.Header().Set(k, v)
		}
		h.ServeHTTP(w, r)
	})
}

// 获取 http.Server 实例，相对于 http 的默认实现，指定了 ErrorLog 字段。
func (s *Server) getServer(port string, h http.Handler) *http.Server {
	srv := &http.Server{
		Addr:         port,
		Handler:      h,
		ErrorLog:     logs.ERROR(),
		ReadTimeout:  s.conf.ReadTimeout * time.Second,
		WriteTimeout: s.conf.WriteTimeout * time.Second,
	}

	s.servers = append(s.servers, srv)

	return srv
}
