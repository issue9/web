// SPDX-License-Identifier: MIT

package app

import (
	"crypto/tls"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/issue9/localeutil"
	"github.com/issue9/mux/v7"
	"golang.org/x/crypto/acme/autocert"
)

type (
	httpConfig struct {
		// 端口
		//
		// 格式与 [http.Server.Addr] 相同。可以为空，表示由 [http.Server] 确定其默认值。
		Port string `yaml:"port,omitempty" json:"port,omitempty" xml:"port,attr,omitempty"`

		// 网站的域名证书
		//
		// NOTE: 不能同时与 ACME 生效
		Certificates []*certificate `yaml:"certificates,omitempty" json:"certificates,omitempty" xml:"certificate,omitempty"`

		// ACME 协议的证书
		//
		// NOTE: 不能同时与 Certificates 生效
		ACME *acme `yaml:"acme,omitempty" json:"acme,omitempty" xml:"acme,omitempty"`

		tlsConfig *tls.Config

		ReadTimeout       duration `yaml:"readTimeout,omitempty" json:"readTimeout,omitempty" xml:"readTimeout,attr,omitempty"`
		WriteTimeout      duration `yaml:"writeTimeout,omitempty" json:"writeTimeout,omitempty" xml:"writeTimeout,attr,omitempty"`
		IdleTimeout       duration `yaml:"idleTimeout,omitempty" json:"idleTimeout,omitempty" xml:"idleTimeout,attr,omitempty"`
		ReadHeaderTimeout duration `yaml:"readHeaderTimeout,omitempty" json:"readHeaderTimeout,omitempty" xml:"readHeaderTimeout,attr,omitempty"`
		MaxHeaderBytes    int      `yaml:"maxHeaderBytes,omitempty" json:"maxHeaderBytes,omitempty" xml:"maxHeaderBytes,attr,omitempty"`

		// 自定义报头功能
		//
		// 报头会输出到包括 404 在内的所有请求返回。可以为空。
		// 报头内容可能会被后续的中间件修改。
		Headers []header `yaml:"headers,omitempty" json:"headers,omitempty" xml:"headers>header,omitempty"`

		// 自定义[跨域请求]设置项
		//
		// [跨域请求]: https://developer.mozilla.org/zh-CN/docs/Web/HTTP/cors
		CORS *cors `yaml:"cors,omitempty" json:"cors,omitempty" xml:"cors,omitempty"`

		routersOptions []mux.Option
	}

	header struct {
		Key   string `yaml:"key" json:"key" xml:"key,attr"`
		Value string `yaml:"val" json:"val" xml:",chardata"`
	}

	certificate struct {
		Cert string `yaml:"cert,omitempty" json:"cert,omitempty" xml:"cert,omitempty"`
		Key  string `yaml:"key,omitempty" json:"key,omitempty" xml:"key,omitempty"`
	}

	acme struct {
		Domains []string `yaml:"domains" json:"domains" xml:"domain"`
		Cache   string   `yaml:"cache" json:"cache" xml:"cache"`
		Email   string   `yaml:"email,omitempty" json:"email,omitempty" xml:"email,omitempty"`

		// 定义提早几天开始续订，如果为 0 表示提早 30 天。
		RenewBefore uint `yaml:"renewBefore,omitempty" json:"renewBefore,omitempty" xml:"renewBefore,attr,omitempty"`
	}

	cors struct {
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

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil || errors.Is(err, fs.ErrExist)
}

func (cert *certificate) sanitize() *ConfigError {
	if !exists(cert.Cert) {
		return &ConfigError{Field: "cert", Message: localeutil.Phrase("%s not found", cert.Cert)}
	}

	if !exists(cert.Key) {
		return &ConfigError{Field: "key", Message: localeutil.Phrase("%s not found", cert.Key)}
	}

	return nil
}

func (h *httpConfig) sanitize() *ConfigError {
	if h.ReadTimeout < 0 {
		return &ConfigError{Field: "readTimeout", Message: localeutil.Phrase("should great than 0")}
	}

	if h.WriteTimeout < 0 {
		return &ConfigError{Field: "writeTimeout", Message: localeutil.Phrase("should great than 0")}
	}

	if h.IdleTimeout < 0 {
		return &ConfigError{Field: "idleTimeout", Message: localeutil.Phrase("should great than 0")}
	}

	if h.ReadHeaderTimeout < 0 {
		return &ConfigError{Field: "readHeaderTimeout", Message: localeutil.Phrase("should great than 0")}
	}

	if h.MaxHeaderBytes < 0 {
		return &ConfigError{Field: "maxHeaderBytes", Message: localeutil.Phrase("should great than 0")}
	}

	if err := h.buildRoutersOptions(); err != nil {
		return err
	}

	return h.buildTLSConfig()
}

func (h *httpConfig) buildHTTPServer(err *log.Logger) *http.Server {
	return &http.Server{
		Addr:              h.Port,
		ReadTimeout:       h.ReadTimeout.Duration(),
		ReadHeaderTimeout: h.ReadHeaderTimeout.Duration(),
		WriteTimeout:      h.WriteTimeout.Duration(),
		IdleTimeout:       h.IdleTimeout.Duration(),
		MaxHeaderBytes:    h.MaxHeaderBytes,
		ErrorLog:          err,
		TLSConfig:         h.tlsConfig,
	}
}

func (h *httpConfig) buildRoutersOptions() *ConfigError {
	opt := make([]mux.Option, 0, 1)

	if len(h.Headers) > 0 {
		opt = append(opt, mux.OnConnection(func(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request) {
			for _, hh := range h.Headers {
				w.Header().Add(hh.Key, hh.Value)
			}
			return w, r
		}))
	}

	if h.CORS != nil {
		c := h.CORS
		opt = append(opt, mux.CORS(c.Origins, c.AllowHeaders, c.ExposedHeaders, c.MaxAge, c.AllowCredentials))
	}

	h.routersOptions = opt
	return nil
}

func (h *httpConfig) buildTLSConfig() *ConfigError {
	if len(h.Certificates) > 0 && h.ACME != nil {
		return &ConfigError{Field: "acme", Message: "不能与 certificates 同时存在"}
	}

	if h.ACME != nil {
		if err := h.ACME.sanitize(); err != nil {
			err.Field = "acme." + err.Field
			return err
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
			return &ConfigError{Field: "certificates", Message: err.Error()}
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	}
	h.tlsConfig = tlsConfig

	return nil
}

func (l *acme) tlsConfig() *tls.Config {
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

func (l *acme) sanitize() *ConfigError {
	if l.Cache == "" || !exists(l.Cache) {
		return &ConfigError{Field: "cache", Message: "不存在该目录或是未指定"}
	}

	if len(l.Domains) == 0 {
		return &ConfigError{Field: "domains", Message: "不能为空"}
	}

	return nil
}

// 封装 time.Duration 以实现对 JSON、XML 和 YAML 的解析
type duration time.Duration

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
