// SPDX-License-Identifier: MIT

// Package cache 缓存接口的定义
package cache

import "github.com/issue9/web/internal/errs"

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
	//
	// NOTE: 获取正确获取由 [Cache.Counter] 设置的值。
	Get(key string, v any) error

	// Set 设置或是添加缓存项
	//
	// key 表示保存该数据的唯一 ID；
	// val 表示保存的数据对象，如果是结构体，需要所有的字段都是公开的或是实现了
	// [Marshaler] 和 [Unmarshaler] 接口，否则在 [Cache.Get] 中将失去这些非公开的字段。
	// ttl 表示过了该时间，缓存项将被回收。如果该值为 0，该值永远不会回收。
	Set(key string, val any, ttl int) error

	// Delete 删除一个缓存项
	Delete(string) error

	// Exists 判断一个缓存项是否存在
	Exists(string) bool

	// Counter 返回计数器操作接口
	//
	// val 和 ttl 表示在该计数器不存在时，初始化的值以及回收时间。
	Counter(key string, val uint64, ttl int) Counter
}

// Counter 计数器需要实现的接口
//
// [Cache] 支持自定义的序列化接口，但是对于自增等纯数值操作，
// 各个缓存服务都实现自有的快捷操作，无法适用自定义的序列化。
//
// 由 Counter 设置的值，也无法由 [Cache.Get] 读取到正确的值，
// 但是可以由 [Cache.Exists]、[Cache.Delete] 和 [Cache.Set] 进行相应的操作。
//
// 各个驱动对自增值的类型定义是不同的，
// 只有在 [0,math.MaxInt32] 范围内的数值是安全的。
//
// NOTE: 只能用于正整数和零。
type Counter interface {
	// Incr 增加计数并返回增加后的值
	Incr(uint64) (uint64, error)

	// Decr 减少数值并返回减少后的值
	Decr(uint64) (uint64, error)

	// Value 返回该计数器的当前值
	//
	// 如果出错将返回默认值。
	Value() (uint64, error)
}

type CleanableCache interface {
	Cache

	// Clean 清除所有的缓存内容
	Clean() error
}

// Driver 所有缓存驱动需要实现的接口
//
// 对于数据的序列化相关操作可直接调用 [caches.Marshal] 和 [caches.Unmarshal]
// 进行处理，如果需要自行处理，需要对实现 [Serializer] 接口的数据进行处理。
//
// 新的驱动可以采用 [cachetest] 对接口进行测试，看是否符合要求。
//
// [cachetest]: https://pkg.go.dev/github.com/issue9/web/cache/cachetest
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
