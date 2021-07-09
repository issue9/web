// SPDX-License-Identifier: MIT

package config

import (
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/logs/v3"
)

func TestWebconfig_buildCache(t *testing.T) {
	a := assert.New(t)
	l, err := logs.New(nil)
	a.NotError(err).NotNil(l)

	cfg := &Webconfig{}
	a.NotError(cfg.buildCache(l))
	a.NotNil(cfg.cache)

	cfg = &Webconfig{Cache: &Cache{DSN: "1h"}}
	a.NotError(cfg.buildCache(l))
	a.NotNil(cfg.cache)

	cfg = &Webconfig{Cache: &Cache{Type: "memory", DSN: "1h"}}
	a.NotError(cfg.buildCache(l))
	a.NotNil(cfg.cache)

	cfg = &Webconfig{Cache: &Cache{Type: "not-exists"}}
	a.Error(cfg.buildCache(l))
	a.Nil(cfg.cache)
}
