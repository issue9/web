// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

// Package jsonp JSONP 序列化操作
package jsonp

import (
	"encoding/json/v2"
	"io"
	"net/http"
	"sync"
	"unique"

	"github.com/issue9/errwrap"
	"github.com/issue9/mux/v10/header"

	"github.com/issue9/web"
	"github.com/issue9/web/codec"
	"github.com/issue9/web/locales"
	"github.com/issue9/web/problems"
)

type contextKeyType int

const contextKey contextKeyType = 1

// ProblemCallbackUnset 表示查询参数中没有 key 指定的参数或是参数值为空时的应该返回的 [web.Problem]
var ProblemCallbackUnset = unique.Make("callback-unset")

var once = &sync.Once{}

// NewJSONP 提供对 JSONP 编码的支持
func NewJSONP() *codec.MediaType { return web.NewMediaType(header.Javascript, "", marshal, unmarshal) }

// Install 安装 JSONP 的处理方式
//
// key 用于指定回函数名称的查询参数名称；
func Install(s *web.Server, key string) {
	if key == "" {
		panic("key 不能为空")
	}

	once.Do(func() {
		s.Vars().Store(contextKey, key)
		s.Problems().Add("", &problems.Item{
			ID:     ProblemCallbackUnset,
			Type:   "callback-unset",
			Status: http.StatusBadRequest,
			Title:  web.Phrase("the jsonp callback is unset"),
		})
	})
}

func marshal(ctx *web.Context, v any) ([]byte, error) {
	if ctx == nil {
		panic("ctx 不能为空")
	}

	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	ctxKey, found := ctx.Server().Vars().Load(contextKey)
	if !found {
		return data, nil
	}

	q, err := ctx.Queries(true)
	if err != nil {
		return data, err
	}

	key := ctxKey.(string)
	callback := q.String(key, "")
	if callback == "" {
		return nil, web.NewProblemError(ProblemCallbackUnset, locales.NewFieldError(key, web.StringPhrase("can not be empty")))
	}

	b := errwrap.StringBuilder{}
	b.WString(callback).WByte('(').WBytes(data).WByte(')')
	return []byte(b.String()), b.Err
}

func unmarshal(r io.Reader, v any) error { return json.UnmarshalRead(r, v) }
