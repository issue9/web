// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

// 此对象不能和 operationRenderer 放在同一文件，且需要在其之前编译，
// 否则会和造成泛型对象相互引用，无法编译。
//
// TODO: https://github.com/golang/go/issues/70230
type pathItemRenderer struct {
	Get        *operationRenderer             `json:"get,omitempty" yaml:"get,omitempty"`
	Put        *operationRenderer             `json:"put,omitempty" yaml:"put,omitempty"`
	Post       *operationRenderer             `json:"post,omitempty" yaml:"post,omitempty"`
	Delete     *operationRenderer             `json:"delete,omitempty" yaml:"delete,omitempty"`
	Options    *operationRenderer             `json:"options,omitempty" yaml:"options,omitempty"`
	Head       *operationRenderer             `json:"head,omitempty" yaml:"head,omitempty"`
	Patch      *operationRenderer             `json:"patch,omitempty" yaml:"patch,omitempty"`
	Trace      *operationRenderer             `json:"trace,omitempty" yaml:"trace,omitempty"`
	Servers    []*serverRenderer              `json:"servers,omitempty" yaml:"servers,omitempty"`
	Parameters []*renderer[parameterRenderer] `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}
