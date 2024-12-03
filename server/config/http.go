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
	"strings"
	"time"

	"github.com/issue9/logs/v7"
	"github.com/issue9/mux/v9/header"
	"golang.org/x/crypto/acme/autocert"

	"github.com/issue9/web"
	"github.com/issue9/web/filter"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/server"
)

type (
	httpConfig struct {
		// 端口
		//
		// 格式与 [http.Server.Addr] 相同。可以为空，表示由 [http.Server] 确定其默认值。
		Port string `yaml:"port,omitempty" json:"port,omitempty" xml:"port,attr,omitempty"`

		// [web.Router.URL] 的默认前缀
		//
		// 如果是非标准端口，应该带上端口号。
		//
		// NOTE: 每个路由可使用 [web.WithURLDomain] 重新定义该值。
		URL string `yaml:"url,omitempty" json:"url,omitempty" xml:"url,omitempty"`

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

		ReadTimeout       Duration `yaml:"readTimeout,omitempty" json:"readTimeout,omitempty" xml:"readTimeout,attr,omitempty"`
		WriteTimeout      Duration `yaml:"writeTimeout,omitempty" json:"writeTimeout,omitempty" xml:"writeTimeout,attr,omitempty"`
		IdleTimeout       Duration `yaml:"idleTimeout,omitempty" json:"idleTimeout,omitempty" xml:"idleTimeout,attr,omitempty"`
		ReadHeaderTimeout Duration `yaml:"readHeaderTimeout,omitempty" json:"readHeaderTimeout,omitempty" xml:"readHeaderTimeout,attr,omitempty"`
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
		// 可以有以下几种值：
		//  - disable 禁用 TRACE 请求；
		//  - body 启用 TRACE，且在返回内容中包含了请求端的 body 内容；
		//  - nobody 启用 TRACE，但是在返回内容中不包含请求端的 body 内容；
		// 默认为 disable。
		//
		// NOTE: 这些设置对所有路径均有效，但会被 [web.Routers.New] 的参数修改。
		Trace string           `yaml:"trace,omitempty" json:"trace,omitempty" xml:"trace,omitempty"`
		trace web.RouterOption // 由 Trace 字段转换而来

		init       func(*server.Options)
		httpServer *http.Server
	}

	// Duration 表示时间段
	//
	// 封装 [time.Duration] 以实现对 JSON、XML 和 YAML 的解析
	Duration time.Duration

	headerConfig struct {
		// 报头名称
		Key string `yaml:"key" json:"key" xml:"key,attr"`

		// 报头对应的值
		Value string `yaml:"value" json:"value" xml:",chardata"`
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

var (
	fileExistsRule = filter.V(func(p string) bool {
		_, err := os.Stat(p)
		return err == nil || errors.Is(err, fs.ErrExist)
	}, locales.NotFound)

	durShouldGreatThan0 = filter.V(func(v Duration) bool { return v >= 0 }, locales.ShouldGreatThan(0))
)

func (cert *certificateConfig) sanitize() *web.FieldError {
	return filter.ToFieldError(
		filter.New("cert", &cert.Cert, fileExistsRule),
		filter.New("key", &cert.Key, fileExistsRule),
	)
}

func (h *httpConfig) sanitize(l *logs.Logs) *web.FieldError {
	err := filter.ToFieldError(
		filter.New("readTimeout", &h.ReadTimeout, durShouldGreatThan0),
		filter.New("writeTimeout", &h.WriteTimeout, durShouldGreatThan0),
		filter.New("idleTimeout", &h.IdleTimeout, durShouldGreatThan0),
		filter.New("readHeaderTimeout", &h.ReadHeaderTimeout, durShouldGreatThan0),
		filter.New("maxHeaderBytes", &h.MaxHeaderBytes, filter.V(func(v int) bool { return v >= 0 }, locales.ShouldGreatThan(0))),
		filter.New("requestID", &h.RequestID, filter.S(func(v *string) {
			if *v == "" {
				*v = header.XRequestID
			}
		})),
		filter.New("trace", &h.Trace, filter.S(func(t *string) {
			if *t == "" {
				*t = "disable"
			}
		}), filter.V(func(t string) bool {
			switch strings.ToLower(t) {
			case "body":
				h.trace = web.WithTrace(true)
			case "nobody":
				h.trace = web.WithTrace(false)
			case "disable":
			default:
				return false
			}
			return true
		}, locales.InvalidValue)),
	)
	if err != nil {
		return err
	}

	if err := h.buildTLSConfig(); err != nil {
		return err
	}

	h.buildInit(l)
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

func (h *httpConfig) buildInit(l *logs.Logs) {
	h.init = func(o *server.Options) {
		if len(h.Headers) > 0 {
			o.Plugins = append(o.Plugins, web.PluginFunc(func(s web.Server) {
				s.Routers().Use(web.MiddlewareFunc(func(next web.HandlerFunc, _, _, _ string) web.HandlerFunc {
					return func(ctx *web.Context) web.Responser {
						for _, hh := range h.Headers {
							ctx.Header().Add(hh.Key, hh.Value)
						}
						return next(ctx)
					}
				}))
			}))
		}

		if h.Recovery > 0 {
			o.RoutersOptions = append(o.RoutersOptions, web.WithRecovery(h.Recovery, l.ERROR()))
		}

		if h.CORS != nil {
			c := h.CORS
			cors := web.WithCORS(c.Origins, c.AllowHeaders, c.ExposedHeaders, c.MaxAge, c.AllowCredentials)
			o.RoutersOptions = append(o.RoutersOptions, cors)
		}

		if h.trace != nil {
			o.RoutersOptions = append(o.RoutersOptions, h.trace)
		}

		if h.URL != "" {
			o.RoutersOptions = append(o.RoutersOptions, web.WithURLDomain(h.URL))
		}
	}
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
	return (&autocert.Manager{
		Cache:       autocert.DirCache(l.Cache),
		Prompt:      autocert.AcceptTOS,
		HostPolicy:  autocert.HostWhitelist(l.Domains...),
		RenewBefore: time.Duration(l.RenewBefore) * 24 * time.Hour,
		Email:       l.Email,
	}).TLSConfig()
}

func (l *acmeConfig) sanitize() *web.FieldError {
	return filter.ToFieldError(
		filter.New("cache", &l.Cache, fileExistsRule),
		filter.New("domains", &l.Domains, filter.V(func(v []string) bool { return len(v) > 0 }, locales.CanNotBeEmpty)),
	)
}

// Duration 转换为标准库的 [time.Duration]
func (d Duration) Duration() time.Duration { return time.Duration(d) }

func (d Duration) MarshalText() ([]byte, error) { return []byte(time.Duration(d).String()), nil }

func (d *Duration) UnmarshalText(b []byte) error {
	v, err := time.ParseDuration(string(b))
	if err == nil {
		*d = Duration(v)
	}
	return err
}
