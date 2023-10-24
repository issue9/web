// SPDX-License-Identifier: MIT

package micro

import "github.com/issue9/web"

type Gateway struct {
	web.Server
	registry Registry
	services []Options
}

// NewGateway 声明网关
func NewGateway(s web.Server, r Registry) *Gateway {
	return &Gateway{
		Server:   s,
		registry: r,
	}
}

func (g *Gateway) Serve() error {
	opts, err := g.registry.Services()
	if err != nil {
		return err
	}

	// TODO

	for _, opt := range opts {
		r := g.NewRouter(opt.Name, opt.Matcher)
		r.Any("{path}", func(ctx *web.Context) web.Responser {
			// TODO
			return ctx.NotImplemented()
		})
	}

	return g.Server.Serve()
}
