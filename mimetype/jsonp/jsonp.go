// SPDX-License-Identifier: MIT

// Package jsonp JSONP 序列化操作
package jsonp

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/issue9/errwrap"

	"github.com/issue9/web"
)

const Mimetype = "application/javascript"

type contextKeyType int

type options struct {
	key          string
	unsetProblem string
}

const contextKey contextKeyType = 1

var once = &sync.Once{}

// Install 安装 JSONP 的处理方式
//
// key 用于指定回函数名称的查询参数名称；
// unsetProblem 表示查询参数中没有 key 指定的参数或是参数值为空时的应该返回的 [web.Problem]
func Install(s web.Server, key, unsetProblem string) {
	if key == "" {
		panic("key 不能为空")
	}

	if unsetProblem == "" {
		panic("unsetProblem 不能为空")
	}

	once.Do(func() {
		s.Vars().Store(contextKey, options{key: key, unsetProblem: unsetProblem})
	})
}

func Marshal(ctx *web.Context, v any) ([]byte, error) {
	if ctx == nil {
		panic("ctx 不能为空")
	}

	data, err := json.Marshal(v)
	if err != nil {
		return nil, ctx.Error(err, web.ProblemNotAcceptable)
	}

	v, found := ctx.Server().Vars().Load(contextKey)
	if !found {
		return data, nil
	}
	o := v.(options)

	q, err := ctx.Queries(true)
	if err != nil {
		return data, err
	}

	callback := q.String(o.key, "")
	if callback == "" {
		return nil, ctx.Problem(o.unsetProblem)
	}

	b := errwrap.StringBuilder{}
	b.WString(callback).WByte('(').WBytes(data).WByte(')')
	return []byte(b.String()), b.Err
}

func Unmarshal(r io.Reader, v any) error { return json.NewDecoder(r).Decode(v) }
