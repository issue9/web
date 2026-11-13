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

// ProblemCallbackUnset 表示查询参数中没有 key 指定的参数或是参数值为空时的应该返回的 [web.Problem]
var ProblemCallbackUnset = unique.Make("callback-unset")

var once = &sync.Once{}

type jsonp string

// New 提供对 JSONP 编码的支持
//
// key 用于指定回调函数在地址中的名称
func New(key string) codec.MediaType { return jsonp(key) }

func (j jsonp) Name() string { return header.Javascript }

func (j jsonp) ProblemName() string { return header.Javascript }

func (j jsonp) Marshal(ctx *web.Context, v any) ([]byte, error) {
	if ctx == nil {
		panic("ctx 不能为空")
	}

	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	q, err := ctx.Queries(true)
	if err != nil {
		return data, err
	}

	callback := q.String(string(j), "")
	if callback == "" {
		return nil, web.NewProblemError(ProblemCallbackUnset, locales.NewFieldError(string(j), web.StringPhrase("can not be empty")))
	}

	b := errwrap.StringBuilder{}
	b.WString(callback).WByte('(').WBytes(data).WByte(')')
	return []byte(b.String()), b.Err
}

func (j jsonp) Unmarshal(r io.Reader, v any) error { return json.UnmarshalRead(r, v) }

// Install 安装 JSONP 需要的基本环境
func Install(s *web.Server) {
	once.Do(func() {
		s.Problems().Add("", &problems.Item{
			ID:     ProblemCallbackUnset,
			Type:   "callback-unset",
			Status: http.StatusBadRequest,
			Title:  web.Phrase("the jsonp callback is unset"),
		})
	})
}
