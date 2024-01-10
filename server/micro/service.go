// SPDX-License-Identifier: MIT

package micro

import (
	"time"

	"github.com/issue9/web"
	"github.com/issue9/web/selector"
	"github.com/issue9/web/server/micro/registry"
)

// service 微服务
type service struct {
	web.Server
	registry registry.Registry
	dreg     registry.DeregisterFunc

	peer selector.Peer
}

// NewService 将 [web.Server] 作为微服务节点
func NewService(s web.Server, r registry.Registry, peer selector.Peer) web.Server {
	return &service{
		Server:   s,
		registry: r,
		peer:     peer,
	}
}

func (s *service) Serve() error {
	dreg, err := s.registry.Register(s.Name(), s.peer)
	if err != nil {
		return err
	}
	s.dreg = dreg

	return s.Server.Serve()
}

func (s *service) Close(shutdown time.Duration) {
	if err := s.dreg(); err != nil {
		s.Logs().ERROR().Error(err)
	}

	s.Server.Close(shutdown)
}
