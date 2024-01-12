// SPDX-License-Identifier: MIT

package server

import (
	"github.com/issue9/web"
	"github.com/issue9/web/server/registry"
)

type gateway struct {
	*httpServer
	registry registry.Registry
	mapper   registry.Mapper
}

// NewGateway 声明微服务的网关
func NewGateway(name, version string, o *Options) (web.Server, error) {
	o, err := sanitizeOptions(o, typeGateway)
	if err != nil {
		err.Path = "Options"
		return nil, err
	}

	g := &gateway{
		registry: o.Registry,
		mapper:   o.Mapper,
	}
	g.httpServer = newHTTPServer(name, version, o, g)
	return g, nil
}

func (g *gateway) Serve() error {
	proxy := g.registry.ReverseProxy(g.mapper)

	r := g.Routers().New("proxy", nil)
	r.Any("{path}", func(ctx *web.Context) web.Responser {
		proxy.ServeHTTP(ctx, ctx.Request())
		return nil
	})

	return g.httpServer.Serve()
}
