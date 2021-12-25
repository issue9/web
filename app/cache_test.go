// SPDX-License-Identifier: MIT

package app

import (
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/logs/v3"
)

func TestWebconfig_buildCache(t *testing.T) {
	a := assert.New(t, false)
	l, err := logs.New(nil)
	a.NotError(err).NotNil(l)

	cfg := &Config{logs: l}
	a.NotError(cfg.buildCache())
	a.NotNil(cfg.cache)

	cfg = &Config{Cache: &Cache{DSN: "1h"}}
	a.NotError(cfg.buildCache())
	a.NotNil(cfg.cache)

	cfg = &Config{Cache: &Cache{Type: "memory", DSN: "1h"}}
	a.NotError(cfg.buildCache())
	a.NotNil(cfg.cache)

	cfg = &Config{Cache: &Cache{Type: "not-exists"}}
	a.Error(cfg.buildCache())
	a.Nil(cfg.cache)
}
