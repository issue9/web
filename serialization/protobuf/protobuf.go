// SPDX-License-Identifier: MIT

// Package protobuf 提供对 Google protocol buffers 的支持
package protobuf

import (
	"google.golang.org/protobuf/proto"

	"github.com/issue9/web/serialization"
)

// Version 当前支持的协议版本号
const Version = "3"

// Mimetype 当前协议的 mimetype 值
const Mimetype = "application/protobuf"

// Marshal 提供对 protobuf 的支持
func Marshal(v interface{}) ([]byte, error) {
	if p, ok := v.(proto.Message); ok {
		return proto.Marshal(p)
	}
	return nil, serialization.ErrUnsupported
}

// Unmarshal 提供对 protobuf 的支持
func Unmarshal(buf []byte, v interface{}) error {
	if p, ok := v.(proto.Message); ok {
		return proto.Unmarshal(buf, p)
	}
	return serialization.ErrUnsupported
}
