// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package server

import (
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/cache/caches/memory"

	"github.com/issue9/web/selector"
)

func TestConfigOf_buildMicro(t *testing.T) {
	a := assert.New(t, false)
	c, _ := memory.New()

	conf := &configOf[empty]{}
	a.NotError(conf.buildMicro(c))

	conf = &configOf[empty]{
		Registry: &registryConfig{
			Type:     "cache",
			Strategy: "random",
			Args:     "1s",
		},
	}
	a.NotError(conf.buildMicro(c))

	conf = &configOf[empty]{
		Registry: &registryConfig{
			Type:     "cache",
			Strategy: "weighted-random",
			Args:     "1s",
		},
		Peer: "https://localhost:8080,5",
	}
	a.NotError(conf.buildMicro(c)).
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
	}
	a.NotError(conf.buildMicro(c)).
		Equal(conf.peer, selector.NewWeightedPeer("https://localhost:8080", 5)).
		NotNil(conf.mapper["s1"])

	conf = &configOf[empty]{
		Registry: &registryConfig{
			Type:     "not-exists",
			Strategy: "random",
			Args:     "1s",
		},
	}
	a.Equal(conf.buildMicro(c).Field, "registry.type")
}

func TestRegistryConfig_build(t *testing.T) {
	a := assert.New(t, false)
	c, _ := memory.New()

	conf := &registryConfig{
		Type:     "cache",
		Strategy: "random",
		Args:     "1s",
	}
	a.NotError(conf.build(c))

	conf = &registryConfig{
		Type:     "not-exists",
		Strategy: "random",
		Args:     "1s",
	}
	err := conf.build(c)
	a.Equal(err.Field, "type")

	conf = &registryConfig{
		Type:     "cache",
		Strategy: "not-exists",
		Args:     "1s",
	}
	err = conf.build(c)
	a.Equal(err.Field, "strategy")
}
