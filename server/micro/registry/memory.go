// SPDX-License-Identifier: MIT

package registry

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/issue9/web"
	"github.com/issue9/web/selector"
)

type memory struct {
	s web.Server
	b func() selector.Selector

	items  map[string]selector.Selector
	locker sync.RWMutex
}

// NewMemory 保存在内存的 [Registry] 实现
func NewMemory(s web.Server, b func() selector.Selector) Registry {
	return &memory{
		s:     s,
		b:     b,
		items: make(map[string]selector.Selector, 10),
	}
}

func (m *memory) Register(name string, p selector.Peer) (DeregisterFunc, error) {
	m.locker.Lock()
	s, found := m.items[name]
	m.locker.Unlock()

	if !found {
		s = m.b()
		m.items[name] = s
	}

	if err := s.Add(p); err != nil {
		return nil, err
	}

	return func() error {
		m.locker.Lock()
		s, found := m.items[name]
		m.locker.Unlock()
		if found {
			return s.Del(p.Addr())
		}
		return nil
	}, nil
}

func (m *memory) Discover(name string) (selector.Selector, error) {
	m.locker.RLock()
	s, found := m.items[name]
	m.locker.RUnlock()

	if found {
		return s, nil
	}
	return nil, web.NewLocaleError("not found service %s", name)
}

func (m *memory) ReverseProxy(ms Mapper) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			for name, match := range ms {
				if !match.Match(r.In, nil) {
					continue
				}

				m.locker.RLock()
				s, found := m.items[name]
				m.locker.RUnlock()
				if !found {
					panic(web.NewError(http.StatusNotFound, web.NewLocaleError("not found service %s", name)))
				}

				route, err := s.Next()
				if err != nil {
					panic(err)
				}

				u, err := url.Parse(route)
				if err != nil {
					panic(err) // Selector 实现得不标准
				}

				r.SetURL(u)
				// r.Out.Host = r.In.Host

				break
			}
		},
	}
}
