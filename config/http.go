// SPDX-License-Identifier: MIT

package config

import (
	"crypto/tls"
	"time"

	"golang.org/x/crypto/acme/autocert"

	"github.com/issue9/web/internal/filesystem"
)

type (
	// HTTP 与 http 请求相关的设置
	HTTP struct {
		// 网站的域名证书
		//
		// 不能同时与 LetsEncrypt 生效
		Certificates []*Certificate `yaml:"certificates,omitempty" json:"certificates,omitempty" xml:"certificates,omitempty"`

		// 配置 Let's Encrypt 证书
		//
		// 不能同时与 Certificates 生效
		LetsEncrypt *LetsEncrypt `yaml:"letsEncrypt,omitempty" json:"letsEncrypt,omitempty" xml:"letsEncrypt,omitempty"`

		tlsConfig *tls.Config

		// 应用于 http.Server 的几个变量
		ReadTimeout       Duration `yaml:"readTimeout,omitempty" json:"readTimeout,omitempty" xml:"readTimeout,attr,omitempty"`
		WriteTimeout      Duration `yaml:"writeTimeout,omitempty" json:"writeTimeout,omitempty" xml:"writeTimeout,attr,omitempty"`
		IdleTimeout       Duration `yaml:"idleTimeout,omitempty" json:"idleTimeout,omitempty" xml:"idleTimeout,attr,omitempty"`
		ReadHeaderTimeout Duration `yaml:"readHeaderTimeout,omitempty" json:"readHeaderTimeout,omitempty" xml:"readHeaderTimeout,attr,omitempty"`
		MaxHeaderBytes    int      `yaml:"maxHeaderBytes,omitempty" json:"maxHeaderBytes,omitempty" xml:"maxHeaderBytes,attr,omitempty"`
	}

	// Certificate 证书管理
	Certificate struct {
		Cert string `yaml:"cert,omitempty" json:"cert,omitempty" xml:"cert,omitempty"`
		Key  string `yaml:"key,omitempty" json:"key,omitempty" xml:"key,omitempty"`
	}

	// LetsEncrypt Let's Encrypt 的相关设置
	LetsEncrypt struct {
		Domains []string `yaml:"domains" json:"domains" xml:"domains"`
		Cache   string   `yaml:"cache" json:"cache" xml:"cache"`
		Email   string   `yaml:"email,omitempty" json:"email,omitempty" xml:"email,omitempty"`

		// 定义提早几天开始续订，如果为 0 表示提早 30 天。
		RenewBefore uint `yaml:"renewBefore,omitempty" json:"renewBefore,omitempty" xml:"renewBefore,attr,omitempty"`
	}
)

func (cert *Certificate) sanitize() *Error {
	if !filesystem.Exists(cert.Cert) {
		return &Error{Field: "cert", Message: "文件不存在"}
	}

	if !filesystem.Exists(cert.Key) {
		return &Error{Field: "key", Message: "文件不存在"}
	}

	return nil
}

func (http *HTTP) sanitize() *Error {
	if http.ReadTimeout < 0 {
		return &Error{Field: "readTimeout", Message: "必须大于等于 0"}
	}

	if http.WriteTimeout < 0 {
		return &Error{Field: "writeTimeout", Message: "必须大于等于 0"}
	}

	if http.IdleTimeout < 0 {
		return &Error{Field: "idleTimeout", Message: "必须大于等于 0"}
	}

	if http.ReadHeaderTimeout < 0 {
		return &Error{Field: "readHeaderTimeout", Message: "必须大于等于 0"}
	}

	if http.MaxHeaderBytes < 0 {
		return &Error{Field: "maxHeaderBytes", Message: "必须大于等于 0"}
	}

	return http.buildTLSConfig()
}

func (http *HTTP) buildTLSConfig() *Error {
	if len(http.Certificates) > 0 && http.LetsEncrypt != nil {
		return &Error{Field: "letsEncrypt", Message: "不能与 certificates 同时存在"}
	}

	if http.LetsEncrypt != nil {
		if err := http.LetsEncrypt.sanitize(); err != nil {
			err.Field = "letsEncrypt." + err.Field
			return err
		}

		http.tlsConfig = http.LetsEncrypt.tlsConfig()
		return nil
	}

	tlsConfig := &tls.Config{Certificates: make([]tls.Certificate, 0, len(http.Certificates))}
	for _, certificate := range http.Certificates {
		if err := certificate.sanitize(); err != nil {
			return err
		}

		cert, err := tls.LoadX509KeyPair(certificate.Cert, certificate.Key)
		if err != nil {
			return &Error{Field: "certificates", Message: err.Error()}
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	}
	http.tlsConfig = tlsConfig

	return nil
}

func (l *LetsEncrypt) tlsConfig() *tls.Config {
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

func (l *LetsEncrypt) sanitize() *Error {
	if l.Cache == "" || !filesystem.Exists(l.Cache) {
		return &Error{Field: "cache", Message: "不存在该目录或是未指定"}
	}

	if len(l.Domains) == 0 {
		return &Error{Field: "domains", Message: "不能为空"}
	}

	return nil
}

// Duration 封装 time.Duration 以实现对 JSON、XML 和 YAML 的解析
type Duration time.Duration

// Duration 转换成 time.Duration
func (d Duration) Duration() time.Duration { return time.Duration(d) }

// MarshalText encoding.TextMarshaler 接口
func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

// UnmarshalText encoding.TextUnmarshaler 接口
func (d *Duration) UnmarshalText(b []byte) error {
	v, err := time.ParseDuration(string(b))
	if err == nil {
		*d = Duration(v)
	}
	return err
}
