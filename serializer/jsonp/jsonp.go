// SPDX-License-Identifier: MIT

// Package jsonp JSONP 序列化操作
package jsonp

import (
	"encoding/json"

	"github.com/issue9/errwrap"

	"github.com/issue9/web"
	"github.com/issue9/web/server"
)

const Mimetype = "application/javascript"

// Install 安装 JSONP 的处理方式
func Install(callbackKey string, s *web.Server) {
	s.OnMarshal(Mimetype, nil, func(ctx *server.Context, data []byte) []byte {
		q, err := ctx.Queries(true)
		if err != nil {
			s.Logs().ERROR().Error(err)
			return data
		}

		callback := q.String(callbackKey, "")
		if callback == "" {
			return data
		}

		b := errwrap.StringBuilder{}
		b.WString(callback).WByte('(').WBytes(data).WByte(')')
		return []byte(b.String())
	})
}

func Marshal(v any) ([]byte, error) { return json.Marshal(v) }

func Unmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }
