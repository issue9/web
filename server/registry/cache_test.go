// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

package registry_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert/v6"

	"github.com/issue9/web"
	"github.com/issue9/web/selector"
	"github.com/issue9/web/server/registry"
)

func TestCache_Discover(t *testing.T) {
	a := assert.New(t, false)
	s, _ := newTestServer(a)

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
	s, _ := newTestServer(a)
	c := registry.NewCache(web.NewCache("registry", s.Cache()), registry.NewRoundRobinStrategy(), time.Second)
	a.NotNil(c)

	t.Run("name=", func(t *testing.T) {
		proxy := c.ReverseProxy("", s)
		a.NotNil(proxy)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		a.Panic(func() { proxy.ServeHTTP(w, r) })
	})

	proxy := c.ReverseProxy("s1", s)
	a.NotNil(proxy)

	t.Run("未注册服务", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		a.Panic(func() { proxy.ServeHTTP(w, r) })
	})

	t.Run("正常", func(t *testing.T) {
		s1, startup1 := newTestServer(a)
		go s1.Serve() // s1 是的启动器 httptest.Server，不需要真正等待其完成
		defer s1.Close(0)

		p1Dreg, err := c.Register("s1", selector.NewPeer(startup1.URL()))
		a.NotError(err).NotNil(p1Dreg)
		time.Sleep(2 * time.Second) // 等待 proxy 更新
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/path", nil)
		proxy.ServeHTTP(w, r)
		a.Equal(w.Result().StatusCode, http.StatusNotFound)
	})
}
