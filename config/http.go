// SPDX-License-Identifier: MIT

package config

import (
	"crypto/tls"
	"net/url"

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

		tlsConfig *tls.Config `yaml:"-" json:"-" xml:"-"`

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
		Domains     []string `yaml:"domains" json:"domains" xml:"domains"`
		Cache       string   `yaml:"cache" json:"cache" xml:"cache"`
		Email       string   `yaml:"email,omitempty" json:"email,omitempty" xml:"email,omitempty"`
		ForceRSA    bool     `yaml:"forceRSA,omitempty" json:"forceRSA,omitempty" xml:"forceRSA,attr,omitempty"`
		RenewBefore Duration `yaml:"renewBefore,omitempty" json:"renewBefore,omitempty" xml:"renewBefore,attr,omitempty"`
	}

	// Debug 调试信息的配置
	Debug struct {
		Pprof string `yaml:"pprof,omitempty" json:"pprof,omitempty" xml:"pprof,omitempty"`
		Vars  string `yaml:"vars,omitempty" json:"vars,omitempty" xml:"vars,omitempty"`
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

func (http *HTTP) sanitize(root *url.URL) *Error {
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

	return http.buildTLSConfig(root)
}

func (http *HTTP) buildTLSConfig(root *url.URL) *Error {
	if root.Scheme == "https" &&
		len(http.Certificates) == 0 &&
		http.LetsEncrypt == nil {
		return &Error{Field: "certificates", Message: "HTTPS 必须指定至少一张证书"}
	}

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
	m := &autocert.Manager{
		Cache:       autocert.DirCache(l.Cache),
		Prompt:      autocert.AcceptTOS,
		HostPolicy:  autocert.HostWhitelist(l.Domains...),
		RenewBefore: l.RenewBefore.Duration(),
		ForceRSA:    l.ForceRSA,
		Email:       l.Email,
	}

	return m.TLSConfig()
}

func (l *LetsEncrypt) sanitize() *Error {
	if l.Cache == "" || !filesystem.Exists(l.Cache) {
		return &Error{Field: "cache", Message: "不存在该目录或是为空"}
	}

	if len(l.Domains) == 0 {
		return &Error{Field: "domains", Message: "不能为空"}
	}

	return nil
}
