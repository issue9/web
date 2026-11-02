// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"testing"

	"github.com/issue9/assert/v6"
)

func TestConfig_buildCache(t *testing.T) {
	a := assert.New(t, false)
	_, f := buildFilter()

	cfg := &configOf[empty]{}
	cfg.buildCache(f)
	a.False(f.HasError()).NotNil(cfg.cache)

	cfg = &configOf[empty]{Cache: &cacheConfig{Type: "memory", DSN: "1h"}}
	cfg.buildCache(f)
	a.False(f.HasError()).NotNil(cfg.cache)

	cfg = &configOf[empty]{Cache: &cacheConfig{Type: "not-exists"}}
	cfg.buildCache(f)
	a.True(f.HasError()).Nil(cfg.cache)
}
