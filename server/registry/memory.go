// SPDX-License-Identifier: MIT

package registry

import (
	"net/http/httputil"
	"sync"

	"github.com/issue9/web"
	"github.com/issue9/web/selector"
)

type memory struct {
	b      func() selector.Selector
	items  map[string]selector.Selector
	locker sync.RWMutex
}

// NewMemory 保存在内存的 [Registry] 实现
func NewMemory(b func() selector.Selector) Registry {
	return &memory{
		b:     b,
		items: make(map[string]selector.Selector, 10),
	}
}

func (m *memory) Register(name string, p selector.Peer) (DeregisterFunc, error) {
	s := m.find(name)
	if s == nil {
		s = m.b()
		m.locker.Lock()
		m.items[name] = s
		m.locker.Unlock()
	}

	if err := s.Add(p); err != nil {
		return nil, err
	}

	return func() error {
		if s := m.find(name); s != nil {
			return s.Del(p.Addr())
		}
		return nil
	}, nil
}

func (m *memory) Discover(name string) (selector.Selector, error) {
	if s := m.find(name); s != nil {
		return s, nil
	}
	return nil, web.NewLocaleError("not found micro service %s", name)
}

func (m *memory) ReverseProxy(ms Mapper) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite: ms.asRewrite(m.find),
	}
}

func (m *memory) find(name string) selector.Selector {
	m.locker.RLock()
	s := m.items[name]
	m.locker.RUnlock()
	return s
}
