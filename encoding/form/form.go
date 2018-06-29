// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package form 用于处理 www-form-urlencoded 编码
package form

import "net/url"

// MimeType 当前编码的媒体类型
const MimeType = "application/x-www-form-urlencoded"

// Marshal 针对 www-form-urlencoded 内容的 MarshalFunc 实现
func Marshal(v interface{}) ([]byte, error) {
	if vals, ok := v.(url.Values); ok {
		return []byte(vals.Encode()), nil
	}

	return nil, nil
}

// Unmarshal 针对 www-form-urlencoded 内容的 UnmarshalFunc 实现
func Unmarshal(data []byte, v interface{}) error {
	vals, err := url.ParseQuery(string(data))
	if err != nil {
		return err
	}

	v = vals
	return nil
}
