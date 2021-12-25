// SPDX-License-Identifier: MIT

package app

import "github.com/issue9/mux/v5"

type Router struct {
	// 是否忽略大小写
	//
	// 如果为 true，那么客户请求的 URL 都将被转换为小写字符。
	// 不会影响服务端添加的路由项。
	CaseInsensitive bool `yaml:"caseInsensitive,omitempty" json:"caseInsensitive,omitempty" xml:"caseInsensitive,attr,omitempty"`

	// 跨域的相关设置
	//
	// 为空表示禁用跨域的相关设置。
	CORS *CORS `yaml:"cors,omitempty" json:"cors,omitempty" xml:"cors,omitempty"`

	options []mux.Option
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
	opts := make([]mux.Option, 0, 2)

	if r.CORS != nil {
		if err := r.CORS.sanitize(); err != nil {
			err.Field = "cors." + err.Field
			return err
		}
		o := r.CORS
		opts = append(opts, mux.CORS(o.Origins, o.AllowHeaders, o.ExposedHeaders, o.MaxAge, o.AllowCredentials))
	}

	if r.CaseInsensitive {
		opts = append(opts, mux.CaseInsensitive)
	}

	r.options = opts
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
