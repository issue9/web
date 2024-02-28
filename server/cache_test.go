// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package server

import (
	"testing"

	"github.com/issue9/assert/v4"
)

func TestConfig_buildCache(t *testing.T) {
	a := assert.New(t, false)

	cfg := &configOf[empty]{}
	a.NotError(cfg.buildCache()).Nil(cfg.cache)

	cfg = &configOf[empty]{Cache: &cacheConfig{Type: "memory", DSN: "1h"}}
	a.NotError(cfg.buildCache()).NotNil(cfg.cache)

	cfg = &configOf[empty]{Cache: &cacheConfig{Type: "not-exists"}}
	a.Error(cfg.buildCache()).Nil(cfg.cache)
}
