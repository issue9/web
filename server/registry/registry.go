// SPDX-License-Identifier: MIT

// Package registry 服务注册与发现
package registry

import (
	"bytes"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/issue9/errwrap"

	"github.com/issue9/web"
	"github.com/issue9/web/selector"
)

var bufferPool = &sync.Pool{New: func() any { return &errwrap.Buffer{} }}

// Registry 服务注册与发现需要实现的接口
type Registry interface {
	// Register 注册服务
	//
	// name 服务名称；
	// peer 节点信息；
	//
	// 返回一个用于注销当前服务的方法；
	Register(name string, peer selector.Peer) (DeregisterFunc, error)

	// Discover 返回指定名称的服务节点
	//
	// name 为服务的名称；
	// s 为调用者关联的 [web.Server] 对象；
	Discover(name string, s web.Server) selector.Selector

	// ReverseProxy 返回反向代理对象
	//
	// s 为调用者关联的 [web.Server] 对象；
	ReverseProxy(m Mapper, s web.Server) *httputil.ReverseProxy
}

type DeregisterFunc = func() error

type Mapper map[string]web.RouterMatcher

// RewriteFunc 转换为 [httputil.ProxyRequest.Rewrite] 字段类型的函数
//
// f 为查找指定名称的 [selector.Selector] 对象。当 [Mapper] 中找到匹配的项时，
// 需要通过 f 找对应的 [selector.Selector] 对象。
func (ms Mapper) RewriteFunc(f func(string) selector.Selector) func(r *httputil.ProxyRequest) {
	return func(r *httputil.ProxyRequest) {
		if len(ms) == 0 {
			panic(web.NewError(http.StatusNotFound, web.NewLocaleError("empty mapper")))
		}

		for name, match := range ms {
			if !match.Match(r.In, nil) {
				continue
			}

			s := f(name)
			if s == nil {
				panic(web.NewError(http.StatusNotFound, web.NewLocaleError("not found micro service %s", name)))
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
			return
		}
	}
}

func marshalPeers(peers []selector.Peer) ([]byte, error) {
	// TODO 改用 JSON 序列化，性能会差一些，但是不用手动用分号进行分隔，减少出错的可能性。

	buf := bufferPool.Get().(*errwrap.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	for _, p := range peers {
		data, err := p.MarshalText()
		if err != nil {
			return nil, err
		}
		buf.WBytes(data)
		buf.WByte(';')
	}

	return buf.Bytes(), buf.Err
}

func unmarshalPeers(n func() selector.Peer, data []byte) ([]selector.Peer, error) {
	items := bytes.Split(data, []byte{';'})

	peers := make([]selector.Peer, 0, len(items))
	for _, item := range items {
		if len(item) == 0 {
			continue
		}

		p := n()
		if err := p.UnmarshalText(item); err != nil {
			return nil, err
		}
		peers = append(peers, p)
	}
	return peers, nil
}
