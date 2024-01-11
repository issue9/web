// SPDX-License-Identifier: MIT

package micro

import (
	"github.com/issue9/web"
	"github.com/issue9/web/server/micro/registry"
)

type gateway struct {
	web.Server
	registry registry.Registry
	mapper   registry.Mapper
}

// NewGateway 声明网关
func NewGateway(s web.Server, r registry.Registry, mapper registry.Mapper) web.Server {
	return &gateway{
		Server:   s,
		registry: r,
		mapper:   mapper,
	}
}

func (g *gateway) Serve() error {
	proxy := g.registry.ReverseProxy(g.mapper)

	r := g.Routers().New("proxy", nil)
	r.Any("{path}", func(ctx *web.Context) web.Responser {
		proxy.ServeHTTP(ctx, ctx.Request())
		return nil
	})

	return g.Server.Serve()
}
