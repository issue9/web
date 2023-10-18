// SPDX-License-Identifier: MIT

// Package jsonp JSONP 序列化操作
package jsonp

import (
	"encoding/json"
	"sync"

	"github.com/issue9/errwrap"

	"github.com/issue9/web"
)

const Mimetype = "application/javascript"

type contextKeyType int

const contextKey contextKeyType = 1

var once = &sync.Once{}

// Install 安装 JSONP 的处理方式
//
// callbackKey 用于指定回函数名称的查询参数名称
func Install(callbackKey string, s web.Server) {
	once.Do(func() {
		s.Vars().Store(contextKey, callbackKey)
	})
}

func BuildMarshal(ctx *web.Context) web.MarshalFunc {
	if ctx == nil {
		return json.Marshal
	}

	return func(v any) ([]byte, error) {
		data, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		key, found := ctx.Server().Vars().Load(contextKey)
		if !found {
			return data, nil
		}

		q, err := ctx.Queries(true)
		if err != nil {
			return data, err
		}

		callback := q.String(key.(string), "")
		if callback == "" {
			return data, nil
		}

		b := errwrap.StringBuilder{}
		b.WString(callback).WByte('(').WBytes(data).WByte(')')
		return []byte(b.String()), nil
	}
}

func Unmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }
