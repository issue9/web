// SPDX-License-Identifier: MIT

package micro

import (
	"time"

	"github.com/issue9/web"
)

// Service 微服务
type Service struct {
	web.Server
	registry Registry
	o        Options
}

// NewService 将 [web.Server] 作为微服务节点
func NewService(s web.Server, r Registry, url string, m web.RouterMatcher) *Service {
	return &Service{
		Server:   s,
		registry: r,
		o: Options{
			Name:    s.Name(),
			ID:      s.UniqueID(),
			URL:     url,
			Matcher: m,
		},
	}
}

func (s *Service) Serve() error {
	if err := s.registry.Register(s.o); err != nil {
		return err
	}
	return s.Server.Serve()
}

func (s *Service) Close(shutdown time.Duration) {
	if err := s.registry.Deregister(s.o); err != nil {
		s.Logs().ERROR().Error(err)
	}

	s.Server.Close(shutdown)
}
