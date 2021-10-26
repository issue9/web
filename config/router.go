// SPDX-License-Identifier: MIT

package config

import "github.com/issue9/mux/v5"

type Router struct {
	// 是否禁用自动生成 HEAD 请求
	DisableHead bool `yaml:"disableHead,omitempty" json:"disableHead,omitempty" xml:"disableHead,attr,omitempty"`

	// 跨域的相关设置
	//
	// 为空表示禁用跨域的相关设置。
	CORS *CORS `yaml:"cors,omitempty" json:"cors,omitempty" xml:"cors,omitempty"`
	cors *mux.CORS
}

// CORS 跨域设置
type CORS struct {
	// Origins 对应 Origin
	//
	// 可以是 *，如果包含了 *，那么其它的设置将不再启作用。
	// 此字段将被用于与请求头的 Origin 字段作验证，以确定是否放行该请求。
	//
	// 如果此值为空，表示不启用跨域的相关设置。
	Origins []string `yaml:"origins,omitempty" json:"origins,omitempty" xml:"origin,omitempty"`

	// AllowHeaders 对应 Access-Control-Allow-Headers
	//
	// 可以包含 *，表示可以是任意值，其它值将不再启作用。
	AllowHeaders []string `yaml:"allowHeaders,omitempty" json:"allowHeaders,omitempty" xml:"allowHeader,omitempty"`

	// ExposedHeaders 对应 Access-Control-Expose-Headers
	ExposedHeaders []string `yaml:"exposedHeaders,omitempty" json:"exposedHeaders,omitempty" xml:"exposedHeader,omitempty"`

	// MaxAge 对应 Access-Control-Max-Age
	//
	// 0 不输出该报头；
	// -1 表示禁用；
	// 其它 >= -1 的值正常输出数值；
	MaxAge int `yaml:"maxAge,omitempty" json:"maxAge,omitempty" xml:"maxAge,attr,omitempty"`

	// AllowCredentials 对应 Access-Control-Allow-Credentials
	AllowCredentials bool `yaml:"allowCredentials,omitempty" json:"allowCredentials,omitempty" xml:"allowCredentials,attr,omitempty"`
}

func (r *Router) sanitize() *Error {
	if r.CORS == nil {
		r.cors = mux.DeniedCORS()
		return nil
	}

	if err := r.CORS.sanitize(); err != nil {
		err.Field = "cors." + err.Field
		return err
	}

	r.cors = &mux.CORS{
		Origins:          r.CORS.Origins,
		AllowHeaders:     r.CORS.AllowHeaders,
		ExposedHeaders:   r.CORS.ExposedHeaders,
		MaxAge:           r.CORS.MaxAge,
		AllowCredentials: r.CORS.AllowCredentials,
	}

	return nil
}

func (c *CORS) sanitize() *Error {
	for _, o := range c.Origins {
		if o == "*" {
			if c.AllowCredentials {
				return &Error{Field: "allowCredentials", Message: "不能与 origins=* 同时成立"}
			}
		}
	}

	if c.MaxAge < -1 {
		return &Error{Field: "maxAge", Message: "必须 >= -1"}
	}

	return nil
}
