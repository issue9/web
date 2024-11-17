// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"encoding/json"

	"golang.org/x/text/message"

	"github.com/issue9/web"
)

// Ref 定义了 $ref
type Ref struct {
	Ref         string
	Summary     web.LocaleStringer
	Description web.LocaleStringer
}

type refRenderer struct {
	Ref         string `json:"$ref" yaml:"$ref"`
	Summary     string `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// 生成可用于渲染的对象
//
// typ 表示在 components 中的名称
func (ref *Ref) build(p *message.Printer, typ string) *refRenderer {
	if ref.Ref == "" {
		panic("ref 不能为空")
	}

	return &refRenderer{
		Ref:         "#/components/" + typ + "/" + ref.Ref,
		Summary:     sprint(p, ref.Summary),
		Description: sprint(p, ref.Description),
	}
}

// 渲染对象
//
// 包含了 $ref 和对象本身，仅在 $ref 不为空的情况下才渲染对象本身，否则只渲染 $ref
type renderer[T any] struct {
	ref *refRenderer
	obj *T
}

func newRenderer[T any](ref *refRenderer, obj *T) *renderer[T] {
	if ref == nil && obj == nil {
		panic("ref 和 obj 不能同时为 nil")
	}

	return &renderer[T]{ref: ref, obj: obj}
}

func (r *renderer[T]) MarshalJSON() ([]byte, error) {
	if r.ref != nil {
		return json.Marshal(r.ref)
	}
	return json.Marshal(r.obj)
}

func (r *renderer[T]) MarshalYAML() (any, error) {
	if r.ref != nil {
		return r.ref, nil
	}
	return r.obj, nil
}
