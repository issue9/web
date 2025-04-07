// SPDX-FileCopyrightText: 2024-2025 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"crypto/md5"
	"encoding/hex"
	sj "encoding/json"
	"strconv"

	sy "github.com/goccy/go-yaml"

	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/html"
	"github.com/issue9/web/mimetype/json"
	"github.com/issue9/web/mimetype/yaml"
)

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
		return sj.Marshal(r.ref)
	}
	return sj.Marshal(r.obj)
}

func (r *renderer[T]) MarshalYAML() ([]byte, error) {
	if r.ref != nil {
		return sy.Marshal(r.ref)
	}
	return sy.Marshal(r.obj)
}

func (o *openAPIRenderer) MarshalHTML() (name string, data any) {
	return o.templateName, o
}

// Handler 创建只包含给定标签的文档接口
//
// 目前支持以下几种格式：
//   - json 通过将 accept 报头设置为 [json.Mimetype] 返回 JSON 格式的数据；
//   - yaml 通过将 accept 报头设置为 [yaml.Mimetype] 返回 YAML 格式的数据；
//   - html 通过将 accept 报头设置为 [html.Mimetype] 返回 HTML 格式的数据。
//     需要通过 [WithHTML] 进行配置，可参考 [github.com/issue9/web/mimetype/html]；
//
// NOTE: 支持的输出格式限定在以上几种，但是最终是否能正常输出以上几种格式，
// 还需要由 [web.Server] 是否配置相应的解码方式。
//
// 如果 tag 不为空，则表示该接口只显示与这些标签关联的文档。
func (d *Document) Handler(tag ...string) web.HandlerFunc {
	return func(ctx *web.Context) web.Responser {
		if d.disable {
			return ctx.NotFound()
		}

		if m := ctx.Mimetype(false); (m != json.Mimetype && m != yaml.Mimetype && m != html.Mimetype) ||
			(m == html.Mimetype && d.templateName == "") {
			return ctx.Problem(web.ProblemNotAcceptable)
		}

		return web.NotModified(func() (string, bool) {
			// 引起 ETag 变化的几个要素
			etag := strconv.Itoa(int(d.last.Unix())) + "/" +
				ctx.Mimetype(false) + "/" +
				ctx.LanguageTag().String()
			h := md5.New()
			h.Write([]byte(etag))
			val := h.Sum(nil)
			return hex.EncodeToString(val), true
		}, func() (any, error) {
			return d.build(ctx.LocalePrinter(), ctx.LanguageTag(), tag), nil
		})
	}
}
