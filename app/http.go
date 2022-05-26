// SPDX-License-Identifier: MIT

package app

import (
	"crypto/tls"
	"os"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

type (
	httpConfig struct {
		// 网站的域名证书
		//
		// 不能同时与 ACME 生效
		Certificates []*certificate `yaml:"certificates,omitempty" json:"certificates,omitempty" xml:"certificate,omitempty"`

		// ACME 协议的证书
		//
		// 不能同时与 Certificates 生效
		ACME *acme `yaml:"acme,omitempty" json:"acme,omitempty" xml:"acme,omitempty"`

		tlsConfig *tls.Config

		// 应用于 http.Server 的几个变量
		ReadTimeout       duration `yaml:"readTimeout,omitempty" json:"readTimeout,omitempty" xml:"readTimeout,attr,omitempty"`
		WriteTimeout      duration `yaml:"writeTimeout,omitempty" json:"writeTimeout,omitempty" xml:"writeTimeout,attr,omitempty"`
		IdleTimeout       duration `yaml:"idleTimeout,omitempty" json:"idleTimeout,omitempty" xml:"idleTimeout,attr,omitempty"`
		ReadHeaderTimeout duration `yaml:"readHeaderTimeout,omitempty" json:"readHeaderTimeout,omitempty" xml:"readHeaderTimeout,attr,omitempty"`
		MaxHeaderBytes    int      `yaml:"maxHeaderBytes,omitempty" json:"maxHeaderBytes,omitempty" xml:"maxHeaderBytes,attr,omitempty"`
	}

	// 证书管理
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
)

func (cert *certificate) sanitize() *ConfigError {
	if !exists(cert.Cert) {
		return &ConfigError{Field: "cert", Message: "文件不存在"}
	}

	if !exists(cert.Key) {
		return &ConfigError{Field: "key", Message: "文件不存在"}
	}

	return nil
}

func (http *httpConfig) sanitize() *ConfigError {
	if http.ReadTimeout < 0 {
		return &ConfigError{Field: "readTimeout", Message: "必须大于等于 0"}
	}

	if http.WriteTimeout < 0 {
		return &ConfigError{Field: "writeTimeout", Message: "必须大于等于 0"}
	}

	if http.IdleTimeout < 0 {
		return &ConfigError{Field: "idleTimeout", Message: "必须大于等于 0"}
	}

	if http.ReadHeaderTimeout < 0 {
		return &ConfigError{Field: "readHeaderTimeout", Message: "必须大于等于 0"}
	}

	if http.MaxHeaderBytes < 0 {
		return &ConfigError{Field: "maxHeaderBytes", Message: "必须大于等于 0"}
	}

	return http.buildTLSConfig()
}

func (http *httpConfig) buildTLSConfig() *ConfigError {
	if len(http.Certificates) > 0 && http.ACME != nil {
		return &ConfigError{Field: "acme", Message: "不能与 certificates 同时存在"}
	}

	if http.ACME != nil {
		if err := http.ACME.sanitize(); err != nil {
			err.Field = "acme." + err.Field
			return err
		}

		http.tlsConfig = http.ACME.tlsConfig()
		return nil
	}

	tlsConfig := &tls.Config{Certificates: make([]tls.Certificate, 0, len(http.Certificates))}
	for _, certificate := range http.Certificates {
		if err := certificate.sanitize(); err != nil {
			return err
		}

		cert, err := tls.LoadX509KeyPair(certificate.Cert, certificate.Key)
		if err != nil {
			return &ConfigError{Field: "certificates", Message: err.Error()}
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
	}
	http.tlsConfig = tlsConfig

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

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil || os.IsExist(err)
}
