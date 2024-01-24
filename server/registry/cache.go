// SPDX-License-Identifier: MIT

package registry

import (
	"errors"
	"net/http/httputil"
	"slices"
	"time"

	"github.com/issue9/cache"

	"github.com/issue9/web"
	"github.com/issue9/web/selector"
)

type cacheRegistry struct {
	c    web.Cache
	s    *Strategy
	freq time.Duration
}

const (
	cacheServicesKey    = ":services"
	cachePeersKeySuffix = ":peers"
)

// NewCache 基于 [web.Cache] 的 [Registry] 实现
//
// freq 表示从缓存系统中获取数据的频率；
func NewCache(c web.Cache, s *Strategy, freq time.Duration) Registry {
	return &cacheRegistry{
		c:    c,
		s:    s,
		freq: freq,
	}
}

func (c *cacheRegistry) Register(name string, p selector.Peer) (DeregisterFunc, error) {
	services := []string{}
	if err := c.c.Get(cacheServicesKey, &services); err != nil && !errors.Is(err, cache.ErrCacheMiss()) {
		return nil, err
	}

	if slices.Index(services, name) < 0 { // 需要更新 services
		services = append(services, name)
		if err := c.c.Set(cacheServicesKey, services, cache.Forever); err != nil {
			return nil, err
		}
	}

	var s string
	if err := c.c.Get(name+cachePeersKeySuffix, &s); err != nil && !errors.Is(err, cache.ErrCacheMiss()) {
		return nil, err
	}

	peers, err := unmarshalPeers(c.s.NewPeer, []byte(s))
	if err != nil {
		return nil, err
	}

	if slices.IndexFunc(peers, func(pp selector.Peer) bool { return pp.Addr() == p.Addr() }) >= 0 { // 已存在
		return c.buildDeregisterFunc(name, p), nil
	}

	data, err := marshalPeers(append(peers, p))
	if err != nil {
		return nil, err
	}

	if err := c.c.Set(name+cachePeersKeySuffix, string(data), cache.Forever); err != nil {
		return nil, err
	}

	return c.buildDeregisterFunc(name, p), nil
}

func (c *cacheRegistry) buildDeregisterFunc(name string, p selector.Peer) DeregisterFunc {
	return func() error {
		var s string
		switch err := c.c.Get(name+cachePeersKeySuffix, &s); {
		case errors.Is(err, cache.ErrCacheMiss()): // 线上为空
			return nil
		case err != nil:
			return err
		}

		peers, err := unmarshalPeers(c.s.NewPeer, []byte(s))
		if err != nil {
			return err
		}

		index := slices.IndexFunc(peers, func(e selector.Peer) bool { return e.Addr() == p.Addr() })
		if index < 0 { // 不存在于线上
			return nil
		}

		data, err := marshalPeers(slices.Delete(peers, index, index+1))
		if err != nil {
			return err
		}

		return c.c.Set(name+cachePeersKeySuffix, string(data), cache.Forever)
	}
}

func (c *cacheRegistry) Discover(name string, s web.Server) selector.Selector {
	ss := c.s.NewSelector()

	job := func(time.Time) error {
		var s string
		switch err := c.c.Get(name+cachePeersKeySuffix, &s); {
		case errors.Is(err, cache.ErrCacheMiss()):
			ss.Update()
			return nil
		case err != nil:
			return err
		}

		peers, err := unmarshalPeers(c.s.NewPeer, []byte(s))
		if err == nil {
			ss.Update(peers...)
		}
		return err
	}
	s.Services().AddTicker(web.Phrase("refresh micro service %s for %s", name, s.Name()), job, c.freq, true, true)

	return ss
}

func (c *cacheRegistry) ReverseProxy(ms Mapper, s web.Server) *httputil.ReverseProxy {
	ss := map[string]selector.Updateable{}

	job := func(time.Time) error {
		services := []string{}
		if err := c.c.Get(cacheServicesKey, &services); err != nil {
			return err
		}

	LOOP:
		for _, name := range services {
			var s string
			switch err := c.c.Get(name+cachePeersKeySuffix, &s); {
			case errors.Is(err, cache.ErrCacheMiss()):
				continue LOOP
			case err != nil:
				return err
			}

			sel, found := ss[name]
			if !found {
				ss[name] = c.s.NewSelector()
				sel = ss[name]
			}

			peers, err := unmarshalPeers(c.s.NewPeer, []byte(s))
			if err != nil {
				return err
			}
			sel.Update(peers...)
		}
		return nil
	}

	s.Services().AddTicker(web.Phrase("refresh micro services for gateway %s", s.Name()), job, c.freq, true, true)

	find := func(name string) selector.Selector { return ss[name] }
	return &httputil.ReverseProxy{
		Rewrite: ms.RewriteFunc(func(name string) selector.Selector { return find(name) }),
	}
}
