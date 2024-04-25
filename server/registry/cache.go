// SPDX-FileCopyrightText: 2018-2024 caixw
//
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
	var s string
	if err := c.c.Get(name, &s); err != nil && !errors.Is(err, cache.ErrCacheMiss()) {
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

	if err := c.c.Set(name, string(data), cache.Forever); err != nil {
		return nil, err
	}

	return c.buildDeregisterFunc(name, p), nil
}

func (c *cacheRegistry) buildDeregisterFunc(name string, p selector.Peer) DeregisterFunc {
	return func() error {
		var s string
		switch err := c.c.Get(name, &s); {
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

		return c.c.Set(name, string(data), cache.Forever)
	}
}

func (c *cacheRegistry) Discover(name string, s web.Server) selector.Selector {
	ss := c.s.NewSelector()

	job := func(time.Time) error {
		var s string
		switch err := c.c.Get(name, &s); {
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

func (c *cacheRegistry) ReverseProxy(name string, s web.Server) *httputil.ReverseProxy {
	sel := c.s.NewSelector()

	job := func(time.Time) error {
		var s string
		if err := c.c.Get(name, &s); err != nil {
			if errors.Is(err, cache.ErrCacheMiss()) { // 如果不是因为 cache miss 导致的 s 为空，则应该被正常处理。
				return nil
			}
			return err
		}

		peers, err := unmarshalPeers(c.s.NewPeer, []byte(s))
		if err == nil {
			sel.Update(peers...)
		}
		return err
	}

	s.Services().AddTicker(web.Phrase("refresh micro services for gateway %s", s.Name()), job, c.freq, false, true)

	// [web.Services.AddTicker] 的 imm 并不是马上执行，而是使任务进入可执行的状态。
	// 所以这里需要手动调用一次 job 刷新微服务列表。
	if err := job(s.Now()); err != nil {
		s.Logs().ERROR().Error(err)
	}

	return &httputil.ReverseProxy{
		Director: Selector2Director(sel),
	}
}
