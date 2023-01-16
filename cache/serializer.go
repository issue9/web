// SPDX-License-Identifier: MIT

package cache

// Serializer 缓存系统存取数据时采用的序列化方法
//
// 如果你存储的对象实现了该接口，那么在存取数据时，会采用此方法将对象进行编解码。
// 否则会采用默认的方法进行编辑码。
//
// 实现 Serializer 可以拥有更高效的转换效率，以及一些默认行为不可实现的功能，
// 比如需要对拥有不可导出的字段进行编解码。
type Serializer interface {
	// MarshalCache 将对象转换成 []byte
	MarshalCache() ([]byte, error)

	// UnmarshalCache 从 []byte 中恢复数据
	UnmarshalCache([]byte) error
}
