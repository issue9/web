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
	serviceStartup struct {
		web.Startup
		registry registry.Registry
		peer     selector.Peer
	}

	gatewayStartup struct {
		web.Startup
		registry registry.Registry
	}
)

// NewHTTP 新建 HTTP 服务
//
// id 应用的 ID；
// version 应用的版本号；
// o 初始化应用的额外参数；
func NewHTTP(id, version string, o *Options) (*web.Server, error) {
	o, err := sanitizeOptions(o, typeHTTP)
	if err != nil {
		return nil, err.AddFieldParent("o")
	}

	return o.newServer(id, version), nil
}

// NewService 声明微服务节点
//
// [Options.Registry] 和 [Options.Peer] 不能为空。
func NewService(id, version string, o *Options) (*web.Server, error) {
	o, err := sanitizeOptions(o, typeService)
	if err != nil {
		return nil, err.AddFieldParent("o")
	}

	o.Startup = &serviceStartup{
		registry: o.Registry,
		peer:     o.Peer,
		Startup:  o.Startup,
	}

	return o.newServer(id, version), nil
}

func (startup *serviceStartup) Serve(s *web.Server) error {
	dreg, err := startup.registry.Register(s.ID(), startup.peer)
	if err != nil {
		return err
	}
	s.OnClose(func() error { return dreg() })

	return startup.Startup.Serve(s)
}

// NewGateway 声明微服务的网关
//
// [Options.Mapper] 和 [Options.Peer] 不能为空。
func NewGateway(id, version string, o *Options) (*web.Server, error) {
	o, err := sanitizeOptions(o, typeGateway)
	if err != nil {
		return nil, err.AddFieldParent("o")
	}

	gs := &gatewayStartup{
		registry: o.Registry,
		Startup:  o.Startup,
	}

	o.Startup = gs
	s := o.newServer(id, version)

	for name, match := range o.Mapper {
		proxy := gs.registry.ReverseProxy(name, s)
		s.Routers().New(name, match).Any("{path}", func(ctx *web.Context) web.Responser {
			proxy.ServeHTTP(ctx, ctx.Request())
			return nil
		})
	}

	return s, nil
}
