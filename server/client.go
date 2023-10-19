// SPDX-License-Identifier: MIT

package server

import (
	"net/http"

	"github.com/issue9/web"
)

// NewClient 采用 [Server] 的编码和压缩方式创建 Client 对象
//
// 参数可参考 [NewClient]。
func (srv *httpServer) NewClient(client *http.Client, url, marshalName string) *web.Client {
	c := srv.Codec()
	m := c.Accept(marshalName)
	return web.NewClient(client, url, marshalName, m.MarshalBuilder()(nil), c.ContentType, c.ContentEncoding, c.AcceptEncodingHeader())
}
