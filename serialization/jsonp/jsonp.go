// SPDX-License-Identifier: MIT

// Package jsonp JSONP 序列化操作
package jsonp

import (
	"encoding/json"

	"github.com/issue9/errwrap"
)

const Mimetype = "application/javascript"

type jsonp struct {
	callback string
	data     interface{}
}

// JSONP 返回 JSONP 对象
//
// 采用与 JSON 相同的解析方式。
// callback 表示回调的函数名称，如果为空，则直接返回 data，
// 在输出时也将被当作普通的 JSON 输出。
func JSONP(callback string, data interface{}) interface{} {
	if callback == "" {
		return data
	}
	return &jsonp{callback: callback, data: data}
}

func (j *jsonp) marshal() ([]byte, error) {
	data, err := json.Marshal(j.data)
	if err != nil {
		return nil, err
	}

	b := errwrap.StringBuilder{}
	b.WString(j.callback).WByte('(').WBytes(data).WByte(')')
	return []byte(b.String()), b.Err
}

// Marshal 输出 JSONP 对象
//
// v 如果是由 JSONP 构建的对象，则返回带 callback 的 js 函数；
// 如果是普通的对象，则采用 json.Marshal 将其转换成普通的 JSON 对象返回；
func Marshal(v interface{}) ([]byte, error) {
	switch obj := v.(type) {
	case *jsonp:
		return obj.marshal()
	default:
		return json.Marshal(v)
	}
}
