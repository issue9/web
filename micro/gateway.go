// SPDX-License-Identifier: MIT

package micro

import "github.com/issue9/web"

type Gateway struct {
	*web.Server
}

type GatewayOptions struct {
	// TODO
}

// NewGateway 声明网关
func NewGateway(s *web.Server, o *GatewayOptions) *Gateway {
	return &Gateway{
		Server: s,
	}
}
