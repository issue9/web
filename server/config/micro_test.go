// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package config

import (
	"testing"

	"github.com/issue9/assert/v6"
	"github.com/issue9/cache/drivers/memory"

	"github.com/issue9/web/selector"
)

func TestConfigOf_buildMicro(t *testing.T) {
	a := assert.New(t, false)
	c := memory.New()

	conf := &configOf[empty]{cache: c}
	e, f := buildFilter()
	conf.buildMicro(f)
	a.False(f.HasError()).Empty(e)

	conf = &configOf[empty]{
		cache: c,
		Registry: &registryConfig{
			Type:     "cache",
			Strategy: "random",
			Args:     "1s",
		},
	}
	e, f = buildFilter()
	conf.buildMicro(f)
	a.False(f.HasError())

	conf = &configOf[empty]{
		Registry: &registryConfig{
			Type:     "cache",
			Strategy: "weighted-random",
			Args:     "1s",
		},
		Peer:  "https://localhost:8080,5",
		cache: c,
	}
	e, f = buildFilter()
	conf.buildMicro(f)
	a.False(f.HasError()).
		Equal(conf.peer, selector.NewWeightedPeer("https://localhost:8080", 5))

	conf = &configOf[empty]{
		Registry: &registryConfig{
			Type:     "cache",
			Strategy: "weighted-random",
			Args:     "1s",
		},
		Peer: "https://localhost:8080,5",
		Mappers: []*mapperConfig{
			{Name: "s1", Matcher: "prefix", Args: ",/s1"},
		},
		cache: c,
	}
	e, f = buildFilter()
	conf.buildMicro(f)
	a.False(f.HasError()).
		Equal(conf.peer, selector.NewWeightedPeer("https://localhost:8080", 5)).
		NotNil(conf.mapper["s1"])

	conf = &configOf[empty]{
		Registry: &registryConfig{
			Type:     "not-exists",
			Strategy: "random",
			Args:     "1s",
		},
		cache: c,
	}
	e, f = buildFilter()
	conf.buildMicro(f)
	a.True(f.HasError()).NotEmpty(e["registry.type"])
}

func TestRegistryConfig_build(t *testing.T) {
	a := assert.New(t, false)
	c := memory.New()

	conf := &registryConfig{
		Type:     "cache",
		Strategy: "random",
		Args:     "1s",
	}
	e, f := buildFilter()
	conf.build(c, f)
	a.False(f.HasError()).Empty(e)

	conf = &registryConfig{
		Type:     "not-exists",
		Strategy: "random",
		Args:     "1s",
	}
	e, f = buildFilter()
	conf.build(c, f)
	a.True(f.HasError()).NotEmpty(e["type"])

	conf = &registryConfig{
		Type:     "cache",
		Strategy: "not-exists",
		Args:     "1s",
	}
	e, f = buildFilter()
	conf.build(c, f)
	a.True(f.HasError()).NotEmpty(e["strategy"])
}
