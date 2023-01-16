// SPDX-License-Identifier: MIT

package caches

import (
	"testing"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/cache"
)

var _ cache.Cache = &memcacheDriver{}

func TestMemcache(t *testing.T) {
	a := assert.New(t, false)

	c := NewMemcache("localhost:11211")
	a.NotNil(c)

	testCache(a, c)
	testObject(a, c)
	testCounter(a, c)

	a.NotError(c.Close())
}

func TestMemcache_Close(t *testing.T) {
	a := assert.New(t, false)

	c := NewMemcache("localhost:11211")
	a.NotNil(c)
	a.NotError(c.Set("key", "val", cache.Forever))
	a.NotError(c.Close())

	c = NewMemcache("localhost:11211")
	a.NotNil(c)
	var val string
	a.NotError(c.Get("key", &val)).Equal(val, "val")
}
