// SPDX-FileCopyrightText: 2018-2026 caixw
//
// SPDX-License-Identifier: MIT

// Package server 提供与服务端实现相关的功能
//
// 目前实现了三种类型的服务端：
//   - [NewHTTP] 构建普通的 HTTP 服务；
//   - [NewGateway] 构建微服务的网关服务；
//   - [NewService] 构建微服务；
package server

import (
	"github.com/issue9/web"
	"github.com/issue9/web/selector"
	"github.com/issue9/web/server/registry"
)

type (
	serviceServer struct {
		web.Server
		registry registry.Registry
		peer     selector.Peer
	}

	gatewayServer struct {
		web.Server
		registry registry.Registry
	}
)

func newHTTPServer(id, version string, o *Options) web.Server {
	srv := o.internalServer(id, version)

	for _, plugin := range o.Plugins { // NOTE: 需要保证在最后
		plugin.Plugin(srv)
	}
	return srv
}

// NewHTTP 新建 HTTP 服务
//
// id 应用的 ID；
// version 应用的版本号；
// o 初始化应用的额外参数；
func NewHTTP(id, version string, o *Options) (web.Server, error) {
	o, err := sanitizeOptions(o, typeHTTP)
	if err != nil {
		return nil, err.AddFieldParent("o")
	}

	return newHTTPServer(id, version, o), nil
}

// NewService 声明微服务节点
//
// [Options.Registry] 和 [Options.Peer] 不能为空。
func NewService(id, version string, o *Options) (web.Server, error) {
	o, err := sanitizeOptions(o, typeService)
	if err != nil {
		return nil, err.AddFieldParent("o")
	}

	return &serviceServer{
		registry: o.Registry,
		peer:     o.Peer,
		Server:   newHTTPServer(id, version, o),
	}, nil
}

func (s *serviceServer) Serve() error {
	dreg, err := s.registry.Register(s.ID(), s.peer)
	if err != nil {
		return err
	}
	s.OnClose(func() error { return dreg() })

	return s.Server.Serve()
}

// NewGateway 声明微服务的网关
//
// [Options.Mapper] 和 [Options.Peer] 不能为空。
func NewGateway(id, version string, o *Options) (web.Server, error) {
	o, err := sanitizeOptions(o, typeGateway)
	if err != nil {
		return nil, err.AddFieldParent("o")
	}

	g := &gatewayServer{
		registry: o.Registry,
		Server:   newHTTPServer(id, version, o),
	}

	for name, match := range o.Mapper {
		proxy := g.registry.ReverseProxy(name, g)
		g.Routers().New(name, match).Any("{path}", func(ctx *web.Context) web.Responser {
			proxy.ServeHTTP(ctx, ctx.Request())
			return nil
		})
	}

	return g, nil
}
