// SPDX-License-Identifier: MIT

// Package cache 缓存接口的定义
package cache

import "github.com/issue9/web/errs"

var (
	errCacheMiss  = errs.NewLocaleError("cache miss")
	errInvalidKey = errs.NewLocaleError("invalid cache key")
)

// Cache 缓存内容的访问接口
type Cache interface {
	// Get 获取缓存项
	//
	// 当前不存在时，返回 [ErrCacheMiss] 错误。
	// key 为缓存项的唯一 ID；
	// v 为缓存写入的地址，应该始终为指针类型；
	Get(key string, v any) error

	// Set 设置或是添加缓存项
	//
	// key 表示保存该数据的唯一 ID；
	// val 表示保存的数据对象，如果是结构体，需要所有的字段都是公开的或是实现了
	// [Marshaler] 和 [Unmarshaler] 接口，否则在 [Cache.Get] 中将失去这些非公开的字段。
	// seconds 表示过了该时间，缓存项将被回收。如果该值为 0，该值永远不会回收。
	Set(key string, val any, seconds int) error

	// Delete 删除一个缓存项
	Delete(key string) error

	// Exists 判断一个缓存项是否存在
	Exists(key string) bool
}

type CleanableCache interface {
	Cache

	// Clean 清除所有的缓存内容
	Clean() error
}

// Driver 所有缓存驱动需要实现的接口
//
// 对于数据的序列化相关操作可以调用 [Marshal] 和 [Unmarshal]
// 进行处理，当然自行处理也可以，如果需要自行处理，
// 需要对 [Marshaler] 和 [Unmarshaler] 接口的数据进行处理。
type Driver interface {
	CleanableCache

	// Close 关闭客户端
	Close() error
}

// ErrCacheMiss 当不存在缓存项时返回的错误
func ErrCacheMiss() error { return errCacheMiss }

// ErrInvalidKey key 的格式无效
//
// 部分适配器对 key 可能是有特殊要求的，
// 比如在文件系统中，可能会不允许在 key 中包含 .. 或是 / 等，碰到此类情况，可返回此错误信息。
func ErrInvalidKey() error { return errInvalidKey }
