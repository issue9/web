// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package form 用于处理 www-form-urlencoded 编码
//
// NOTE: 仅支持将数据从 net/url.Values 导出或是写入到 net/url.Values
//
//  func read(w http.ResponseWriter, r *http.Request) {
//      ctx := web.New(w, r)
//      vals := urls.Values{}
//      !ctx.Read(vals)
//  }
//
//  func write(w http.ResponseWriter, r *http.Request) {
//      ctx := web.New(w, r)
//      vals := urls.Values{}
//      vals.Add("name", "caixw")
//      ctx.Render(http.StatusOK, vals, nil)
//  }
package form

import (
	"errors"
	"net/url"
)

var errInvalidType = errors.New("当前只支持 net/url.Values 类型")

// MimeType 当前编码的媒体类型
const MimeType = "application/x-www-form-urlencoded"

// Marshal 针对 www-form-urlencoded 内容的 MarshalFunc 实现
func Marshal(v interface{}) ([]byte, error) {
	if vals, ok := v.(url.Values); ok {
		return []byte(vals.Encode()), nil
	}

	return nil, errInvalidType
}

// Unmarshal 针对 www-form-urlencoded 内容的 UnmarshalFunc 实现
func Unmarshal(data []byte, v interface{}) error {
	vals, err := url.ParseQuery(string(data))
	if err != nil {
		return err
	}

	obj, ok := v.(url.Values)
	if !ok {
		return errInvalidType
	}

	for k, v := range vals {
		obj[k] = v
	}

	return nil
}
