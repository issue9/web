// SPDX-License-Identifier: MIT

// Package jsonp JSONP 序列化操作
package jsonp

import (
	"encoding/json"

	"github.com/issue9/errwrap"
)

const Mimetype = "application/javascript"

type jsonp struct {
	Callback string
	Data     interface{}
}

// JSONP 返回 JSONP 对象
//
// 采用与 JSON 相同的解析方式。
func JSONP(callback string, data interface{}) interface{} {
	return &jsonp{Callback: callback, Data: data}
}

func (j *jsonp) marshal() ([]byte, error) {
	data, err := json.Marshal(j.Data)
	if err != nil {
		return nil, err
	}
	if j.Callback == "" {
		return data, nil
	}

	b := errwrap.StringBuilder{}
	b.WString(j.Callback).WByte('(').WBytes(data).WByte(')')
	return []byte(b.String()), b.Err
}

func Marshal(v interface{}) ([]byte, error) {
	switch obj := v.(type) {
	case *jsonp:
		return obj.marshal()
	default:
		return json.Marshal(v)
	}
}
