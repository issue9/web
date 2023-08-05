// SPDX-License-Identifier: MIT

package app

import (
	"github.com/issue9/config"
	"github.com/issue9/localeutil"

	"github.com/issue9/web/server"
)

var problemFactory = map[string]server.BuildProblemFunc{}

type problemConfig struct {
	// 指定生成 problem 对象的方法
	//
	// 这些名称由 [RegisterProblemBuilder] 注册。当前可用的值有：
	//  - rfc7807
	Builder string `json:"builder,omitempty" xml:"builder,omitempty" yaml:"builder,omitempty"`

	// 指定代码代码的 ID 前缀
	IDPrefix string `json:"idPrefix,omitempty" xml:"idPrefix,omitempty" yaml:"idPrefix,omitempty"`
}

func (p *problemConfig) sanitize() (*server.Problems, *config.FieldError) {
	if p == nil {
		return nil, nil
	}

	ps := &server.Problems{IDPrefix: p.IDPrefix}

	if p.Builder != "" {
		f, found := problemFactory[p.Builder]
		if !found {
			return nil, config.NewFieldError("builder", localeutil.Error("%s not found", p.Builder))
		}
		ps.Builder = f
	}

	return ps, nil
}

// RegisterProblemBuilder 注册用于生成 [server.Problem] 对象的方法
//
// 如果存在同名，则会覆盖。
func RegisterProblemBuilder(name string, b server.BuildProblemFunc) {
	problemFactory[name] = b
}

func init() {
	RegisterProblemBuilder("rfc7807", server.RFC7807Builder)
}
