// SPDX-License-Identifier: MIT

package server

import (
	"testing"

	"github.com/issue9/assert/v3"
)

func TestConfig_buildCache(t *testing.T) {
	a := assert.New(t, false)

	cfg := &configOf[empty]{}
	a.NotError(cfg.buildCache())
	a.Nil(cfg.cache)

	cfg = &configOf[empty]{Cache: &cacheConfig{Type: "memory", DSN: "1h"}}
	a.NotError(cfg.buildCache())
	a.NotNil(cfg.cache)

	cfg = &configOf[empty]{Cache: &cacheConfig{Type: "not-exists"}}
	a.Error(cfg.buildCache())
	a.Nil(cfg.cache)
}