// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package web

import (
	"errors"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"

	"github.com/issue9/context"
	"github.com/issue9/handlers"
	"github.com/issue9/logs"
)

const (
	httpPort  = ":80"
	httpsPort = ":443"
)

// 当启用 HTTPS 时，对 80 端口的处理方式。
const (
	HTTPStateDisabled = iota // 禁止监听 80 端口
	HTTPStateListen          // 监听 80 端口，与 HTTPS 相同的方式处理
	HTTPStateRedirect        // 监听 80 端口，并重定向到 HTTPS
	httpStateSize
)

// 启动 Run() 函数的相关参数。
type Config struct {
	HTTPS        bool                 `json:"https,omitempty"`        // 是否启用 HTTPS
	HTTPState    int                  `json:"httpState,omitempty"`    // 80 端口的状态，仅在 HTTPS 为 true 时，启作用
	CertFile     string               `json:"certFile,omitempty"`     // 当 https 为 true 时，此值为必填
	KeyFile      string               `json:"keyFile,omitempty"`      // 当 https 为 true 时，此值为必填
	Port         string               `json:"port,omitempty"`         // 端口，不指定，默认为 80 或是 443
	Headers      map[string]string    `json:"headers,omitempty"`      // 附加的头信息，头信息可能在其它地方被修改
	Pprof        string               `json:"pprof,omitempty"`        // 指定 pprof 地址
	Static       map[string]string    `json:"static,omitempty"`       // 静态内容，键名为 URL 路径，键值为文件地址
	ReadTimeout  time.Duration        `json:"readTimeout,omitempty"`  // http.Server.ReadTimeout 的值，单位：秒
	WriteTimeout time.Duration        `json:"writeTimeout,omitempty"` // http.Server.WriteTimeout 的值，单位：秒
	ErrHandler   handlers.RecoverFunc `json:"-"`                      // 错误处理
	Before       http.Handler         `json:"-"`                      // 所有路由之前执行的内容
	After        http.Handler         `json:"-"`                      // 所有路由之后执行的内容
}

// 检测 cfg 的各项字段是否合法，
// 初始化相关内容。比如静态文件路由等。
func (cfg *Config) init() error {
	if len(cfg.Port) == 0 {
		if cfg.HTTPS {
			cfg.Port = httpsPort
		} else {
			cfg.Port = httpPort
		}
	} else if cfg.Port[0] != ':' {
		cfg.Port = ":" + cfg.Port
	}

	if cfg.HTTPState < 0 || cfg.HTTPState >= httpStateSize {
		return errors.New("无效的httpState值")
	}

	// 在其它之前调用
	return cfg.buildStaticModule()
}

// 构建一个静态文件服务模块
func (cfg *Config) buildStaticModule() error {
	if len(cfg.Static) == 0 {
		return nil
	}

	m, err := NewModule("web-static")
	if err != nil {
		return err
	}

	for url, dir := range cfg.Static {
		m.Get(url, http.StripPrefix(url, handlers.Compress(http.FileServer(http.Dir(dir)))))
	}

	return nil
}

// 构建一个从 HTTP 跳转到 HTTPS 的路由服务。
func (cfg *Config) httpRedirectListenAndServe() error {
	srv := getServer(cfg, httpPort, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := r.URL
		url.Host = r.Host + cfg.Port
		url.Scheme = "HTTPS"

		urlStr := url.String()
		Info(r, "301 HTTP==>HTTPS:", urlStr)
		http.Redirect(w, r, urlStr, http.StatusMovedPermanently)
	}))

	return srv.ListenAndServe()
}

func (cfg *Config) buildHandler(h http.Handler) http.Handler {
	h = cfg.buildBefore(h)

	h = cfg.buildHeader(h)

	h = cfg.buildAfter(h)

	// 清理 context 的相关内容
	h = context.FreeHandler(h)

	// 错误处理
	h = handlers.Recovery(h, cfg.ErrHandler)

	// 在最外层添加调试地址，保证调试内容不会被其它 handler 干扰。
	return cfg.buildPprof(h)
}

// 根据 config.Pprof 决定是否包装调试地址
func (cfg *Config) buildPprof(h http.Handler) http.Handler {
	if len(cfg.Pprof) > 0 {
		logs.Debug("web:", "开启了调试功能，地址为：", cfg.Pprof)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, cfg.Pprof) {
				h.ServeHTTP(w, r)
				return
			}

			path := r.URL.Path[len(cfg.Pprof):]
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
		})
	}

	return h
}

func (cfg *Config) buildBefore(h http.Handler) http.Handler {
	if cfg.Before == nil {
		return h
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.Before.ServeHTTP(w, r)
		h.ServeHTTP(w, r)
	})
}

func (cfg *Config) buildAfter(h http.Handler) http.Handler {
	if cfg.After == nil {
		return h
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
		cfg.After.ServeHTTP(w, r)
	})
}

func (cfg *Config) buildHeader(h http.Handler) http.Handler {
	if len(cfg.Headers) <= 0 {
		return h
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range cfg.Headers {
			w.Header().Set(k, v)
		}
		h.ServeHTTP(w, r)
	})
}

// 运行路由，执行监听程序。
func (cfg *Config) run() error {
	if err := cfg.init(); err != nil {
		return err
	}

	h := cfg.buildHandler(serveMux)

	if cfg.HTTPS {
		switch cfg.HTTPState {
		case HTTPStateListen:
			logs.Infof("开始监听%v端口", httpPort)
			go getServer(cfg, httpPort, h).ListenAndServe()
		case HTTPStateRedirect:
			logs.Infof("开始监听%v端口", httpPort)
			go cfg.httpRedirectListenAndServe()
		}

		logs.Infof("开始监听%v端口", cfg.Port)
		return getServer(cfg, cfg.Port, h).ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
	}

	logs.Infof("开始监听%v端口", cfg.Port)
	return getServer(cfg, cfg.Port, h).ListenAndServe()
}

// 获取 http.Server 实例，相对于 http 的默认实现，指定了 ErrorLog 字段。
func getServer(cfg *Config, port string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:         port,
		Handler:      h,
		ErrorLog:     logs.ERROR(),
		ReadTimeout:  cfg.ReadTimeout * time.Second,
		WriteTimeout: cfg.WriteTimeout * time.Second,
	}
}
