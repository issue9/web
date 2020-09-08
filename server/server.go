// SPDX-License-Identifier: MIT

// Package server 核心功能的实现
package server

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/issue9/mux/v2"
	"github.com/issue9/scheduled"

	wctx "github.com/issue9/web/context"
	"github.com/issue9/web/internal/webconfig"
)

// Server 程序运行实例
type Server struct {
	http.Server

	builder   *wctx.Builder
	uptime    time.Time
	services  []*Service
	scheduled *scheduled.Server
	webConfig *webconfig.WebConfig
	modules   []*Module

	// 当 shutdown 延时关闭时，通过此事件确定 Serve() 的返回时机。
	closed chan struct{}
}

// New 声明一个新的 Server 实例
func New(conf *webconfig.WebConfig, builder *wctx.Builder) (*Server, error) {
	srv := &Server{
		Server: http.Server{
			Addr:              ":" + strconv.Itoa(conf.Port),
			ErrorLog:          builder.Logs().ERROR(),
			ReadTimeout:       conf.ReadTimeout.Duration(),
			WriteTimeout:      conf.WriteTimeout.Duration(),
			IdleTimeout:       conf.IdleTimeout.Duration(),
			ReadHeaderTimeout: conf.ReadHeaderTimeout.Duration(),
			MaxHeaderBytes:    conf.MaxHeaderBytes,
			Handler:           builder.Handler(),
		},
		builder:   builder,
		uptime:    time.Now(),
		services:  make([]*Service, 0, 100),
		scheduled: scheduled.NewServer(conf.Location),
		webConfig: conf,
		closed:    make(chan struct{}, 1),
	}

	for url, dir := range conf.Static {
		h := http.StripPrefix(url, http.FileServer(http.Dir(dir)))
		srv.builder.Router().Get(url+"{path}", h)
	}

	srv.AddService(srv.scheduledService, "计划任务")

	// 加载固有的中间件，需要在 srv 初始化之后调用
	srv.buildMiddlewares(conf)

	if conf.Plugins != "" {
		if err := srv.loadPlugins(conf.Plugins); err != nil {
			return nil, err
		}
	}

	return srv, nil
}

// Builder 管理返回给客户端的错误信息
func (srv *Server) Builder() *wctx.Builder {
	return srv.builder
}

// Uptime 启动的时间
//
// 时区信息与配置文件中的相同
func (srv *Server) Uptime() time.Time {
	return srv.uptime
}

// Mux 返回相关的 mux.Mux 实例
func (srv *Server) Mux() *mux.Mux {
	return srv.builder.Router().Mux()
}

// IsDebug 是否处于调试模式
func (srv *Server) IsDebug() bool {
	return srv.webConfig.Debug
}

// Path 生成路径部分的地址
//
// 基于 app.webConfig.URL 中的路径部分。
func (srv *Server) Path(p string) string {
	p = srv.webConfig.URLPath + p
	if p != "" && p[0] != '/' {
		p = "/" + p
	}

	return p
}

// URL 构建一条基于 app.webconfig.URL 的完整 URL
func (srv *Server) URL(path string) string {
	if len(path) == 0 {
		return srv.webConfig.URL
	}

	if path[0] != '/' {
		path = "/" + path
	}
	return srv.webConfig.URL + path
}

// Run 执行监听程序
//
// 当调用 Shutdown 关闭服务时，会等待其完成未完的服务，才返回 http.ErrServerClosed
func (srv *Server) Run() (err error) {
	conf := srv.webConfig

	srv.runServices()

	if !conf.HTTPS {
		err = srv.ListenAndServe()
	} else {
		cfg := &tls.Config{}
		for _, certificate := range conf.Certificates {
			cert, err := tls.LoadX509KeyPair(certificate.Cert, certificate.Key)
			if err != nil {
				return err
			}
			cfg.Certificates = append(cfg.Certificates, cert)
		}
		cfg.BuildNameToCertificate()

		srv.TLSConfig = cfg
		err = srv.ListenAndServeTLS("", "")
	}

	// 由 Shutdown() 或 Close() 主动触发的关闭事件，才需要等待其执行完成，
	// 其它错误直接返回，否则一些内部错误会永远卡在此处无法返回。
	if errors.Is(err, http.ErrServerClosed) {
		<-srv.closed
	}
	return err
}

// Close 关闭服务
//
// 无论配置文件如果设置，此函数都是直接关闭服务，不会等待。
func (srv *Server) Close() error {
	defer func() {
		srv.stopServices()
		srv.closed <- struct{}{}
	}()

	return srv.Server.Close()
}

// Shutdown 关闭所有服务
//
// 根据配置文件中的配置项，决定当前是直接关闭还是延时之后关闭。
func (srv *Server) Shutdown(ctx context.Context) error {
	defer func() {
		srv.stopServices()
		srv.closed <- struct{}{}
	}()

	err := srv.Server.Shutdown(ctx)
	if err != nil && err != context.DeadlineExceeded {
		return err
	}
	return nil
}

// Location 当前设置的时区信息
func (srv *Server) Location() *time.Location {
	return srv.webConfig.Location
}
