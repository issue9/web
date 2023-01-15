// SPDX-License-Identifier: MIT

package caches

import (
	"testing"
	"time"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web/cache"
)

var _ cache.Cache = &memoryDriver{}

func TestMemory(t *testing.T) {
	a := assert.New(t, false)

	c := NewMemory(500 * time.Millisecond)
	a.NotNil(c)

	testCache(a, c)
	testObject(a, c)

	a.NotError(c.Close())
}
