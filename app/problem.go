// SPDX-License-Identifier: MIT

package app

import (
	"github.com/issue9/web/internal/errs"
	"github.com/issue9/web/server"
)

var problemFactory = map[string]server.BuildProblemFunc{}

type Problem struct {
	// 指定生成 Problem 对象的方法
	//
	// 这些名称由 [RegisterProblemBuilder] 注册。
	Builder     string `json:"builder,omitempty" xml:"builder,omitempty" yaml:"builder,omitempty"`
	builderFunc server.BuildProblemFunc

	// 指定代码代码的 ID 前缀
	IDPrefix string `json:"idPrefix,omitempty" xml:"idPrefix,omitempty" yaml:"idPrefix,omitempty"`
}

func (p *Problem) sanitize() (*server.Problems, *errs.FieldError) {
	if p == nil {
		return nil, nil
	}

	ps := &server.Problems{IDPrefix: p.IDPrefix}

	if p.Builder != "" {
		f, found := problemFactory[p.Builder]
		if !found {
			return nil, errs.NewFieldError("builder", errs.NewLocaleError("%s not found", p.Builder))
		}
		ps.Builder = f
	}

	return ps, nil
}

// RegisterProblemBuilder 注册用于生成 Problem 对象的方法
//
// 如果存在同名，则会覆盖。
func RegisterProblemBuilder(name string, b server.BuildProblemFunc) {
	problemFactory[name] = b
}

func init() {
	RegisterProblemBuilder("rfc7807", server.RFC7807Builder)
}
