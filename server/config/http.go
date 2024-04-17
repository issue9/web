// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"crypto/tls"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/issue9/logs/v7"
	"github.com/issue9/mux/v8"
	"github.com/issue9/mux/v8/header"
	"golang.org/x/crypto/acme/autocert"

	"github.com/issue9/web"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/server"
)

type (
	httpConfig struct {
		// 端口
		//
		// 格式与 [http.Server.Addr] 相同。可以为空，表示由 [http.Server] 确定其默认值。
		Port string `yaml:"port,omitempty" json:"port,omitempty" xml:"port,attr,omitempty"`

		// x-request-id 的报头名称
		//
		// 如果为空，则采用 [header.XRequestID] 作为默认值。
		RequestID string `yaml:"requestID,omitempty" json:"requestID,omitempty" xml:"requestID,omitempty"`

		// 网站的域名证书
		//
		// NOTE: 不能同时与 ACME 生效
		Certificates []*certificateConfig `yaml:"certificates,omitempty" json:"certificates,omitempty" xml:"certificates>certificate,omitempty"`

		// ACME 协议的证书
		//
		// NOTE: 不能同时与 Certificates 生效
		ACME *acmeConfig `yaml:"acme,omitempty" json:"acme,omitempty" xml:"acme,omitempty"`

		tlsConfig *tls.Config

		ReadTimeout       duration `yaml:"readTimeout,omitempty" json:"readTimeout,omitempty" xml:"readTimeout,attr,omitempty"`
		WriteTimeout      duration `yaml:"writeTimeout,omitempty" json:"writeTimeout,omitempty" xml:"writeTimeout,attr,omitempty"`
		IdleTimeout       duration `yaml:"idleTimeout,omitempty" json:"idleTimeout,omitempty" xml:"idleTimeout,attr,omitempty"`
		ReadHeaderTimeout duration `yaml:"readHeaderTimeout,omitempty" json:"readHeaderTimeout,omitempty" xml:"readHeaderTimeout,attr,omitempty"`
		MaxHeaderBytes    int      `yaml:"maxHeaderBytes,omitempty" json:"maxHeaderBytes,omitempty" xml:"maxHeaderBytes,attr,omitempty"`

		// Recovery 拦截 panic 时反馈给客户端的状态码
		//
		// NOTE: 这些设置对所有路径均有效，但会被 [web.Routers.New] 的参数修改。
		Recovery int `yaml:"recovery,omitempty" json:"recovery,omitempty" xml:"recovery,attr,omitempty"`

		// 自定义报头功能
		//
		// 报头会输出到包括 404 在内的所有请求返回。可以为空。
		//
		// NOTE: 如果是与 CORS 相关的定义，则可能在 CORS 字段的定义中被修改。
		//
		// NOTE: 报头内容可能会被后续的中间件修改。
		Headers []headerConfig `yaml:"headers,omitempty" json:"headers,omitempty" xml:"headers>header,omitempty"`

		// 自定义[跨域请求]设置项
		//
		// NOTE: 这些设置对所有路径均有效，但会被 [web.Routers.New] 的参数修改。
		//
		// [跨域请求]: https://developer.mozilla.org/zh-CN/docs/Web/HTTP/cors
		CORS *corsConfig `yaml:"cors,omitempty" json:"cors,omitempty" xml:"cors,omitempty"`

		// Trace 是否启用 TRACE 请求
		//
		// 默认为 false。
		//
		// NOTE: 这些设置对所有路径均有效，但会被 [web.Routers.New] 的参数修改。
		Trace bool `yaml:"trace,omitempty" json:"trace,omitempty" xml:"trace,omitempty"`

		init       func(*server.Options)
		httpServer *http.Server
	}

	headerConfig struct {
		// 报头名称
		Key string `yaml:"key" json:"key" xml:"key,attr"`

		// 报头对应的值
		Value string `yaml:"val" json:"val" xml:",chardata"`
	}

	certificateConfig struct {
		// 公钥文件地址
		Cert string `yaml:"cert,omitempty" json:"cert,omitempty" xml:"cert,omitempty"`

		// 私钥文件地址
		Key string `yaml:"key,omitempty" json:"key,omitempty" xml:"key,omitempty"`
	}

	acmeConfig struct {
		// 申请的域名列表
		Domains []string `yaml:"domains" json:"domains" xml:"domain"`

		// acme 缓存目录
		Cache string `yaml:"cache" json:"cache" xml:"cache"`

		// 申请者邮箱
		Email string `yaml:"email,omitempty" json:"email,omitempty" xml:"email,omitempty"`

		// 定义提早几天开始续订，如果为 0 表示提早 30 天。
		RenewBefore uint `yaml:"renewBefore,omitempty" json:"renewBefore,omitempty" xml:"renewBefore,attr,omitempty"`
	}

	corsConfig struct {
		// 指定跨域中的 Access-Control-Allow-Origin 报头内容
		//
		// 如果为空，表示禁止跨域请示，如果包含了 *，表示允许所有。
		Origins []string `yaml:"origins,omitempty" json:"origins,omitempty" xml:"origins>origin,omitempty"`

		// 表示 Access-Control-Allow-Headers 报头内容
		AllowHeaders []string `yaml:"allowHeaders,omitempty" json:"allowHeaders,omitempty" xml:"allowHeaders>header,omitempty"`

		// 表示 Access-Control-Expose-Headers 报头内容
		ExposedHeaders []string `yaml:"exposedHeaders,omitempty" json:"exposedHeaders,omitempty" xml:"exposedHeaders>header,omitempty"`

		// 表示 Access-Control-Max-Age 报头内容
		//
		// 有以下几种取值：
		//   - 0 不输出该报头，默认值；
		//   - -1 表示禁用；
		//   - 其它 >= -1 的值正常输出数值；
		MaxAge int `yaml:"maxAge,omitempty" json:"maxAge,omitempty" xml:"maxAge,attr,omitempty"`

		// 表示 Access-Control-Allow-Credentials 报头内容
		AllowCredentials bool `yaml:"allowCredentials,omitempty" json:"allowCredentials,omitempty" xml:"allowCredentials,attr,omitempty"`
	}
)

func (conf *configOf[T]) buildHTTP() *web.FieldError {
	if conf.HTTP == nil {
		conf.HTTP = &httpConfig{}
	}
	if err := conf.HTTP.sanitize(conf.Logs.logs); err != nil {
		return err.AddFieldParent("http")
	}
	if conf.HTTP.init != nil {
		conf.init = append(conf.init, conf.HTTP.init)
	}

	return nil
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil || errors.Is(err, fs.ErrExist)
}

func (cert *certificateConfig) sanitize() *web.FieldError {
	if !exists(cert.Cert) {
		return web.NewFieldError("cert", locales.ErrNotFound())
	}

	if !exists(cert.Key) {
		return web.NewFieldError("key", locales.ErrNotFound())
	}

	return nil
}

func (h *httpConfig) sanitize(l *logs.Logs) *web.FieldError {
	if h.ReadTimeout < 0 {
		return web.NewFieldError("readTimeout", locales.ShouldGreatThan(0))
	}

	if h.WriteTimeout < 0 {
		return web.NewFieldError("writeTimeout", locales.ShouldGreatThan(0))
	}

	if h.IdleTimeout < 0 {
		return web.NewFieldError("idleTimeout", locales.ShouldGreatThan(0))
	}

	if h.ReadHeaderTimeout < 0 {
		return web.NewFieldError("readHeaderTimeout", locales.ShouldGreatThan(0))
	}

	if h.MaxHeaderBytes < 0 {
		return web.NewFieldError("maxHeaderBytes", locales.ShouldGreatThan(0))
	}

	if h.RequestID == "" {
		h.RequestID = header.XRequestID
	}

	if err := h.buildTLSConfig(); err != nil {
		return err
	}

	h.init = func(o *server.Options) {
		if len(h.Headers) > 0 {
			o.Init = append(o.Init, func(s web.Server) {
				s.Routers().Use(web.MiddlewareFunc(func(next web.HandlerFunc) web.HandlerFunc {
					return func(ctx *web.Context) web.Responser {
						for _, hh := range h.Headers {
							ctx.Header().Add(hh.Key, hh.Value)
						}
						return next(ctx)
					}
				}))
			})
		}

		h.buildRoutersOptions(o, l)
	}

	h.buildHTTPServer()
	return nil
}

func (h *httpConfig) buildHTTPServer() {
	h.httpServer = &http.Server{
		Addr:              h.Port,
		ReadTimeout:       h.ReadTimeout.Duration(),
		ReadHeaderTimeout: h.ReadHeaderTimeout.Duration(),
		WriteTimeout:      h.WriteTimeout.Duration(),
		IdleTimeout:       h.IdleTimeout.Duration(),
		MaxHeaderBytes:    h.MaxHeaderBytes,
		TLSConfig:         h.tlsConfig,
	}
}

func (h *httpConfig) buildRoutersOptions(o *server.Options, l *logs.Logs) {
	opt := make([]web.RouterOption, 0, 3)

	if h.Recovery > 0 {
		opt = append(opt, web.Recovery(h.Recovery, l.ERROR()))
	}

	if h.CORS != nil {
		c := h.CORS
		opt = append(opt, mux.CORS(c.Origins, c.AllowHeaders, c.ExposedHeaders, c.MaxAge, c.AllowCredentials))
	}

	opt = append(opt, mux.Trace(h.Trace))

	o.RoutersOptions = append(o.RoutersOptions, opt...)
}

func (h *httpConfig) buildTLSConfig() *web.FieldError {
	if len(h.Certificates) > 0 && h.ACME != nil {
		return web.NewFieldError("acme", web.Phrase("conflict with certificates"))
	}

	if h.ACME != nil {
		if err := h.ACME.sanitize(); err != nil {
			return err.AddFieldParent("acme")
		}

		h.tlsConfig = h.ACME.tlsConfig()
		return nil
	}

	tlsConfig := &tls.Config{Certificates: make([]tls.Certificate, 0, len(h.Certificates))}
	for _, certificate := range h.Certificates {
		if err := certificate.sanitize(); err != nil {
			return err
		}

		cert, err := tls.LoadX509KeyPair(certificate.Cert, certificate.Key)
		if err != nil {
			return web.NewFieldError("certificates", err)
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	}
	h.tlsConfig = tlsConfig

	return nil
}

func (l *acmeConfig) tlsConfig() *tls.Config {
	const day = 24 * time.Hour

	m := &autocert.Manager{
		Cache:       autocert.DirCache(l.Cache),
		Prompt:      autocert.AcceptTOS,
		HostPolicy:  autocert.HostWhitelist(l.Domains...),
		RenewBefore: time.Duration(l.RenewBefore) * day,
		Email:       l.Email,
	}

	return m.TLSConfig()
}

func (l *acmeConfig) sanitize() *web.FieldError {
	if l.Cache == "" || !exists(l.Cache) {
		return web.NewFieldError("cache", locales.InvalidValue)
	}

	if len(l.Domains) == 0 {
		return web.NewFieldError("domains", locales.CanNotBeEmpty)
	}

	return nil
}

// 表示时间段，等同于 [time.Duration]
type duration time.Duration // 封装 time.Duration 以实现对 JSON、XML 和 YAML 的解析

func (d duration) Duration() time.Duration { return time.Duration(d) }

func (d duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

func (d *duration) UnmarshalText(b []byte) error {
	v, err := time.ParseDuration(string(b))
	if err == nil {
		*d = duration(v)
	}
	return err
}
