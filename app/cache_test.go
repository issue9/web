// SPDX-License-Identifier: MIT

package app

import (
	"log"
	"testing"

	"github.com/issue9/assert/v2"
)

func TestWebconfig_buildCache(t *testing.T) {
	a := assert.New(t, false)

	cfg := &configOf[empty]{}
	a.NotError(cfg.buildCache(log.Default()))
	a.NotNil(cfg.cache)

	cfg = &configOf[empty]{Cache: &cacheConfig{DSN: "1h"}}
	a.NotError(cfg.buildCache(log.Default()))
	a.NotNil(cfg.cache)

	cfg = &configOf[empty]{Cache: &cacheConfig{Type: "memory", DSN: "1h"}}
	a.NotError(cfg.buildCache(log.Default()))
	a.NotNil(cfg.cache)

	cfg = &configOf[empty]{Cache: &cacheConfig{Type: "not-exists"}}
	a.Error(cfg.buildCache(log.Default()))
	a.Nil(cfg.cache)
}
