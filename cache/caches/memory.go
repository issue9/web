// SPDX-License-Identifier: MIT

package caches

import (
	"strconv"
	"sync"
	"time"

	"github.com/issue9/scheduled"

	"github.com/issue9/web/cache"
)

type memoryDriver struct {
	items *sync.Map
}

type memoryCounter struct {
	driver    *memoryDriver
	key       string
	val       []byte
	originVal uint64
	expires   time.Duration
	locker    sync.RWMutex
}

type item struct {
	val    []byte
	dur    time.Duration
	expire time.Time // 过期的时间
}

func (i *item) update(val any) error {
	bs, err := Marshal(val)
	if err != nil {
		return err
	}
	i.val = bs
	i.expire = time.Now().Add(i.dur)
	return nil
}

func (i *item) isExpired(now time.Time) bool {
	return i.dur != 0 && i.expire.Before(now)
}

// NewMemory 声明一个内存缓存
//
// JobFunc 表示执行内存的操作。
func NewMemory() (cache.Driver, scheduled.JobFunc) {
	mem := &memoryDriver{
		items: &sync.Map{},
	}
	return mem, mem.gc
}

func (d *memoryDriver) Get(key string, v any) error {
	if item, found := d.findItem(key); found {
		return Unmarshal(item.val, v)
	}
	return cache.ErrCacheMiss()
}

func (d *memoryDriver) findItem(key string) (*item, bool) {
	i, found := d.items.Load(key)
	if !found {
		return nil, false
	}
	return i.(*item), true
}

func (d *memoryDriver) Set(key string, val any, ttl time.Duration) error {
	i, found := d.findItem(key)
	if !found {
		bs, err := Marshal(val)
		if err != nil {
			return err
		}

		d.items.Store(key, &item{
			val:    bs,
			dur:    ttl,
			expire: time.Now().Add(ttl),
		})
		return nil
	}

	return i.update(val)
}

func (d *memoryDriver) Delete(key string) error {
	d.items.Delete(key)
	return nil
}

func (d *memoryDriver) Exists(key string) bool {
	_, found := d.items.Load(key)
	return found
}

func (d *memoryDriver) Clean() error {
	d.items.Range(func(key, val any) bool {
		d.items.Delete(key)
		return true
	})
	return nil
}

func (d *memoryDriver) Close() error { return d.Clean() }

func (d *memoryDriver) gc(now time.Time) error {
	d.items.Range(func(key, val any) bool {
		if v := val.(*item); v.isExpired(now) {
			d.items.Delete(key)
		}
		return true
	})
	return nil
}

func (d *memoryDriver) Counter(key string, val uint64, ttl time.Duration) cache.Counter {
	return &memoryCounter{
		driver:    d,
		key:       key,
		val:       []byte(strconv.FormatUint(val, 10)),
		originVal: val,
		expires:   ttl,
	}
}

func (c *memoryCounter) Incr(n uint64) (uint64, error) {
	c.locker.Lock()
	defer c.locker.Unlock()

	v, err := c.init()
	if err != nil {
		return 0, err
	}

	v += n
	c.driver.items.Store(c.key, &item{
		val:    []byte(strconv.FormatUint(v, 10)),
		dur:    c.expires,
		expire: time.Now().Add(c.expires),
	})
	return v, nil
}

func (c *memoryCounter) Decr(n uint64) (uint64, error) {
	c.locker.Lock()
	defer c.locker.Unlock()

	v, err := c.init()
	if err != nil {
		return 0, err
	}
	if n > v {
		v = 0
	} else {
		v -= n
	}
	c.driver.items.Store(c.key, &item{
		val:    []byte(strconv.FormatUint(v, 10)),
		dur:    c.expires,
		expire: time.Now().Add(c.expires),
	})
	return v, nil
}

func (c *memoryCounter) init() (uint64, error) {
	ret, loaded := c.driver.items.LoadOrStore(c.key, &item{
		val:    c.val,
		dur:    c.expires,
		expire: time.Now().Add(c.expires),
	})

	if !loaded {
		return c.originVal, nil
	}
	return strconv.ParseUint(string(ret.(*item).val), 10, 64)
}

func (c *memoryCounter) Value() (uint64, error) {
	c.locker.RLock()
	defer c.locker.RUnlock()

	if item, found := c.driver.findItem(c.key); found {
		return strconv.ParseUint(string(item.val), 10, 64)
	}
	return c.originVal, cache.ErrCacheMiss()
}

func (c *memoryCounter) Delete() error { return c.driver.Delete(c.key) }
