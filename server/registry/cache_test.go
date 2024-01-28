// SPDX-License-Identifier: MIT

package registry_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert/v3"

	"github.com/issue9/web"
	"github.com/issue9/web/selector"
	"github.com/issue9/web/server/registry"
	"github.com/issue9/web/server/servertest"
)

func TestCache_Discover(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)

	c := registry.NewCache(web.NewCache("registry", s.Cache()), registry.NewRandomStrategy(), time.Second)
	a.NotNil(c)

	// 空的
	sel := c.Discover("s1", s)
	addr, err := sel.Next()
	a.Equal(err, selector.ErrNoPeer()).Empty(addr)

	// 注册微服务 s1

	p1Dreg, err := c.Register("s1", selector.NewPeer("http://localhost:8080"))
	a.NotError(err).NotNil(p1Dreg)
	time.Sleep(2 * time.Second) // 等待 Discover 刷新
	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "http://localhost:8080")
	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "http://localhost:8080")

	p2Dreg, err := c.Register("s1", selector.NewPeer("http://localhost:8081"))
	a.NotError(err).NotNil(p2Dreg)
	time.Sleep(2 * time.Second) // 等待 Discover 刷新
	addr, err = sel.Next()
	a.NotError(err).Contains([]string{"http://localhost:8080", "http://localhost:8081"}, addr)
	addr, err = sel.Next()
	a.NotError(err).Contains([]string{"http://localhost:8080", "http://localhost:8081"}, addr)

	a.NotError(p1Dreg())
	time.Sleep(2 * time.Second) // 等待 Discover 刷新
	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "http://localhost:8081")
	addr, err = sel.Next()
	a.NotError(err).Equal(addr, "http://localhost:8081")
}

func TestCache_ReverseProxy(t *testing.T) {
	a := assert.New(t, false)
	s := newTestServer(a)

	c := registry.NewCache(web.NewCache("registry", s.Cache()), registry.NewRoundRobinStrategy(), time.Second)
	a.NotNil(c)

	// 空的

	proxy := c.ReverseProxy("", s)
	a.NotNil(proxy)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	a.Panic(func() { proxy.ServeHTTP(w, r) })

	// 空服务

	proxy = c.ReverseProxy("s1", s)
	a.NotNil(proxy)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	a.Panic(func() { proxy.ServeHTTP(w, r) })

	// 注册微服务 s1

	s1 := newTestServer(a)
	defer servertest.Run(a, s1)()
	defer s1.Close(0)

	p1Dreg, err := c.Register("s1", selector.NewPeer("http://localhost:8080"))
	a.NotError(err).NotNil(p1Dreg)
	time.Sleep(2 * time.Second) // 等待 proxy
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	proxy.ServeHTTP(w, r)
	a.Equal(w.Result().StatusCode, http.StatusNotFound)
}
