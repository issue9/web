// SPDX-License-Identifier: MIT

package cache

import (
	"bytes"
	"encoding/gob"
)

// Marshaler 缓存系统保存数据时采用的序列化方法
//
// 该接口不是必须的，默认会采用 gob 作为序列化方法。
type Marshaler interface {
	MarshalCache() ([]byte, error)
}

// Marshaler 缓存系统读取数据时采用的序列化方法
//
// 该接口不是必须的，默认会采用 gob 作为序列化方法。
type Unmarshaler interface {
	UnmarshalCache([]byte) error
}

// Marshal 序列化对象
//
// 优先查看 v 是否实现了 [Marshaler] 接口，如果未实现，
// 则采用 gob 格式序列化。
func Marshal(v any) ([]byte, error) {
	if m, ok := v.(Marshaler); ok {
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
	if u, ok := v.(Unmarshaler); ok {
		return u.UnmarshalCache(bs)
	}
	return gob.NewDecoder(bytes.NewBuffer(bs)).Decode(v)
}
