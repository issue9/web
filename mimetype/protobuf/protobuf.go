// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package protobuf 提供对 Google protocol buffers 的支持
package protobuf

import (
	"errors"

	"github.com/golang/protobuf/proto"
)

// Version 当前支持的协议版本号
const Version = "3"

// MimeType 当前协议的 mimetype 值
const MimeType = "application/octet-stream"

var errInvalidType = errors.New("无效的类型，只能是 protof.Message")

// Marshal 提供对 protobuf 的支持
func Marshal(v interface{}) ([]byte, error) {
	if p, ok := v.(proto.Message); ok {
		return proto.Marshal(p)
	}

	return nil, errInvalidType
}

// Unmarshal 提供对 protobuf 的支持
func Unmarshal(buf []byte, v interface{}) error {
	if p, ok := v.(proto.Message); ok {
		return proto.Unmarshal(buf, p)
	}

	return errInvalidType
}
