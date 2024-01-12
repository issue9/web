// SPDX-License-Identifier: MIT

package server

import (
	"time"

	"github.com/issue9/web"
	"github.com/issue9/web/selector"
	"github.com/issue9/web/server/registry"
)

type service struct {
	*httpServer
	registry registry.Registry
	dreg     registry.DeregisterFunc
	peer     selector.Peer
}

// NewService 将 [web.Server] 作为微服务节点
func NewService(name, version string, o *Options) (web.Server, error) {
	o, err := sanitizeOptions(o, typeService)
	if err != nil {
		err.Path = "Options"
		return nil, err
	}

	s := &service{
		registry: o.Registry,
		peer:     o.Peer,
	}
	s.httpServer = newHTTPServer(name, version, o, s)
	return s, nil
}

func (s *service) Serve() error {
	dreg, err := s.registry.Register(s.Name(), s.peer)
	if err != nil {
		return err
	}
	s.dreg = dreg

	return s.httpServer.Serve()
}

func (s *service) Close(shutdown time.Duration) {
	if err := s.dreg(); err != nil {
		s.Logs().ERROR().Error(err)
	}
	s.httpServer.Close(shutdown)
}
