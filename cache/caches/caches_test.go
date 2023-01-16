// SPDX-License-Identifier: MIT

package caches

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/cache"
)

func testCounter(a *assert.Assertion, d cache.Driver) {
	c := d.Counter("v1", 50, 1)
	a.NotNil(c)
	v1, err := c.Incr(5)
	a.NotError(err).Equal(v1, 55)
	v1, err = c.Decr(3)
	a.NotError(err).Equal(v1, 52)

	a.True(d.Exists("v1"))
}

// 测试 Cache 基本功能
func testCache(a *assert.Assertion, c cache.Driver) {
	var v string
	err := c.Get("not_exists", &v)
	a.ErrorIs(err, cache.ErrCacheMiss(), "找到了一个并不存在的值").
		Zero(v, "查找一个并不存在的值，且有返回。")

	a.NotError(c.Set("k1", 123, cache.Forever))
	var num int
	err = c.Get("k1", &num)
	a.NotError(err, "Forever 返回未知错误 %s", err).
		Equal(num, 123, "无法正常获取 k1 的值 v1=%d,v2=%d", v, 123)

	// 重新设置 k1
	a.NotError(c.Set("k1", uint(789), 60))
	var unum uint
	err = c.Get("k1", &unum)
	a.NotError(err, "1*time.Hour 的值 k1 返回错误信息 %s", err).
		Equal(unum, 789, "无法正常获取重新设置之后 k1 的值 v1=%d, v2=%d", v, 789)

	// 被 delete 删除
	a.NotError(c.Delete("k1"))
	err = c.Get("k1", &unum)
	a.Equal(err, cache.ErrCacheMiss(), "k1 并未被回收").
		Zero(v, "被删除之后值并未为空：%+v", v)

	// 超时被回收
	a.NotError(c.Set("k1", 123, 1))
	a.NotError(c.Set("k2", 456, 1))
	a.NotError(c.Set("k3", 789, 1))
	time.Sleep(2 * time.Second)
	a.False(c.Exists("k1"), "k1 超时且未被回收")
	a.False(c.Exists("k2"), "k2 超时且未被回收")
	a.False(c.Exists("k3"), "k3 超时且未被回收")

	// Clean
	a.NotError(c.Set("k1", 123, 1))
	a.NotError(c.Set("k2", 456, 1))
	a.NotError(c.Set("k3", 789, 1))
	a.NotError(c.Clean())
	a.False(c.Exists("k1"), "clear 之后 k1 依然存在").
		False(c.Exists("k2"), "clear 之后 k2 依然存在").
		False(c.Exists("k3"), "clear 之后 k3 依然存在")
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

func testObject(a *assert.Assertion, c cache.Driver) {
	obj := &object{Name: "test", age: 5}
	obj2 := &object{Name: "test", age: 5}

	a.NotError(c.Set("obj", obj, cache.Forever))
	var v object
	err := c.Get("obj", &v)
	a.NotError(err).Equal(&v, obj2)
}
