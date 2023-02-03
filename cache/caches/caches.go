// SPDX-License-Identifier: MIT

// Package caches 内置的缓存接口实现
package caches

import (
	"bytes"
	"encoding/gob"

	"github.com/issue9/web/cache"
)

// Marshal 序列化对象
//
// 这是 [cache.Cache] 存储对象时的转换方法，
// 除了判断 [cache.Serializer] 之外，还提供了默认的编码方式。
//
// 大部分时候 [cache.Driver] 的实现者直接调用此方法即可，
// 如果需要自己实现，需要注意 [cache.Serializer] 接口的判断。
func Marshal(v any) ([]byte, error) {
	if m, ok := v.(cache.Serializer); ok {
		return m.MarshalCache()
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Unmarshal(bs []byte, v any) error {
	if u, ok := v.(cache.Serializer); ok {
		return u.UnmarshalCache(bs)
	}
	return gob.NewDecoder(bytes.NewBuffer(bs)).Decode(v)
}
