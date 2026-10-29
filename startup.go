// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

package web

import (
	"context"
	"errors"
	"net/http"
)

// Startup 服务的启动器
type Startup interface {
	// Init 初始化启动器
	//
	// 只有调用了此方法之后，接口的其它方法才是有效果的。
	Init(h http.Handler)

	// Serve 启动服务
	Serve(s *Server) error

	// 关闭服务
	Close(context.Context) error
}

type httpServerStartup struct {
	hs *http.Server
}

// NewHTTPServerStartup 声明一个基于 [http.Server] 的启动器
func NewHTTPServerStartup(hs *http.Server) Startup {
	return &httpServerStartup{hs: hs}
}

func (b *httpServerStartup) Init(h http.Handler) { b.hs.Handler = h }

func (b *httpServerStartup) Serve(s *Server) error {
	if c := b.hs.TLSConfig; c != nil && (len(c.Certificates) > 0 || c.GetCertificate != nil) {
		return b.hs.ListenAndServeTLS("", "")
	} else {
		return b.hs.ListenAndServe()
	}
}

func (b *httpServerStartup) Close(ctx context.Context) error {
	if err := b.hs.Shutdown(ctx); err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}
