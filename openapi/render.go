// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package openapi

import (
	"crypto/md5"
	"encoding/hex"
	sj "encoding/json"
	"slices"
	"strconv"
	"strings"

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

func (r *renderer[T]) MarshalYAML() (any, error) {
	if r.ref != nil {
		return r.ref, nil
	}
	return r.obj, nil
}

type documentQuery struct {
	Tags []string `query:"tag"`
}

// Handler 实现 [web.HandlerFunc] 接口
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
// 该路由接受 tag 查询参数，在未指定参数的情况下，表示返回所有接口，
// 如果指定了参数，则只返回带指定标签的接口，多个标签以逗号分隔。
func (d *Document) Handler(ctx *web.Context) web.Responser {
	if d.disable {
		return ctx.NotImplemented()
	}

	q := &documentQuery{}
	if resp := ctx.QueryObject(true, q, web.ProblemBadRequest); resp != nil {
		return resp
	}

	if m := ctx.Mimetype(false); (m != json.Mimetype && m != yaml.Mimetype && m != html.Mimetype) ||
		(m == html.Mimetype && d.templateName == "") {
		return ctx.Problem(web.ProblemNotAcceptable)
	}

	dataURL := ctx.Request().URL.Path
	if len(q.Tags) > 0 {
		dataURL += "?tag=" + strings.Join(q.Tags, ",")
	}

	return web.NotModified(func() (string, bool) {
		slices.Sort(q.Tags)

		// 引起 ETag 变化的几个要素
		etag := strconv.Itoa(int(d.last.Unix())) + "/" +
			strings.Join(q.Tags, ",") + "/" +
			ctx.Mimetype(false) + "/" +
			ctx.LanguageTag().String()
		h := md5.New()
		h.Write([]byte(etag))
		val := h.Sum(nil)
		return hex.EncodeToString(val), true
	}, func() (any, error) {
		return d.build(ctx.LocalePrinter(), ctx.LanguageTag(), q.Tags), nil
	})
}

func (o *openAPIRenderer) MarshalHTML() (name string, data any) {
	return o.templateName, o
}
