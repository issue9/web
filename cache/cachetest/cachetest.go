// SPDX-License-Identifier: MIT

// Package cachetest 缓存的测试用例
package cachetest

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/cache"
)

// Counter 测试计数器
func Counter(a *assert.Assertion, d cache.Driver) {
	c := d.Counter("v1", 50, time.Second)
	a.NotNil(c)

	v1, err := c.Value()
	a.ErrorIs(err, cache.ErrCacheMiss()).Equal(v1, 50)

	v1, err = c.Incr(5)
	a.NotError(err).Equal(v1, 55)
	v1, err = c.Value()
	a.Nil(err).Equal(v1, 55)

	v1, err = c.Decr(3)
	a.NotError(err).Equal(v1, 52)
	v1, err = c.Value()
	a.Nil(err).Equal(v1, 52)

	a.True(d.Exists("v1"))

	c.Delete()
	a.False(d.Exists("v1"))

	// 没有值的情况 Decr
	c = d.Counter("v2", 50, time.Second)
	a.NotNil(c)
	a.NotError(c.Delete())
	v2, err := c.Decr(3)
	a.NotError(err).Equal(v2, 47)
	v2, err = c.Value()
	a.Nil(err).Equal(v2, 47)

	v2, err = c.Decr(4888)
	a.NotError(err).Equal(v2, 0)

	v2, err = c.Value()
	a.NotError(err).Equal(v2, 0)

	v2, err = c.Decr(47)
	a.NotError(err).Equal(v2, 0)

	// 多个 Counter 指向同一个 key

	c1 := d.Counter("v3", 50, time.Second)
	c2 := d.Counter("v3", 50, time.Second)

	v1, err = c1.Decr(5)
	a.NotError(err).Equal(v1, 45)
	v2, err = c2.Value()
	a.NotError(err).Equal(v2, 45)
}

// Basic 测试基本功能
func Basic(a *assert.Assertion, c cache.Driver) {
	var v string
	err := c.Get("not_exists", &v)
	a.ErrorIs(err, cache.ErrCacheMiss(), "找到了一个并不存在的值").
		Zero(v, "查找一个并不存在的值，且有返回。")

	a.NotError(c.Set("k1", 123, cache.Forever))
	var num int
	err = c.Get("k1", &num)
	a.NotError(err, "Forever 返回未知错误 %s", err).
		Equal(num, 123, "无法正常获取 k1 的值 v1=%s,v2=%d", v, 123)

	// 重新设置 k1
	a.NotError(c.Set("k1", uint(789), time.Minute))
	var unum uint
	err = c.Get("k1", &unum)
	a.NotError(err, "1*time.Hour 的值 k1 返回错误信息 %s", err).
		Equal(unum, 789, "无法正常获取重新设置之后 k1 的值 v1=%s, v2=%d", v, 789)

	// 被 delete 删除
	a.NotError(c.Delete("k1"))
	err = c.Get("k1", &unum)
	a.Equal(err, cache.ErrCacheMiss(), "k1 并未被回收").
		Zero(v, "被删除之后值并未为空：%+v", v)

	// 超时被回收
	a.NotError(c.Set("k1", 123, time.Second))
	a.NotError(c.Set("k2", 456, time.Second))
	a.NotError(c.Set("k3", 789, time.Second))
	time.Sleep(2 * time.Second)
	a.False(c.Exists("k1"), "k1 超时且未被回收")
	a.False(c.Exists("k2"), "k2 超时且未被回收")
	a.False(c.Exists("k3"), "k3 超时且未被回收")

	// Clean
	a.NotError(c.Set("k1", 123, time.Second))
	a.NotError(c.Set("k2", 456, time.Second))
	a.NotError(c.Set("k3", 789, time.Second))
	a.NotError(c.Clean())
	a.False(c.Exists("k1"), "clean 之后 k1 依然存在").
		False(c.Exists("k2"), "clean 之后 k2 依然存在").
		False(c.Exists("k3"), "clean 之后 k3 依然存在")
}

type object struct {
	Name string
	age  int
}

func (o *object) MarshalCache() ([]byte, error) {
	return []byte(o.Name + "," + strconv.Itoa(o.age)), nil
}

func (o *object) UnmarshalCache(bs []byte) error {
	fields := strings.Split(string(bs), ",")
	if len(fields) != 2 {
		return errors.New("error")
	}

	o.Name = fields[0]
	age, err := strconv.Atoi(fields[1])
	if err != nil {
		return err
	}
	o.age = age

	return nil
}

// Object 测试对象的缓存
func Object(a *assert.Assertion, c cache.Driver) {
	obj := &object{Name: "test", age: 5}
	obj2 := &object{Name: "test", age: 5}

	a.NotError(c.Set("obj", obj, cache.Forever))
	var v object
	err := c.Get("obj", &v)
	a.NotError(err).Equal(&v, obj2)
}
