// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

// Package servertest 提供测试 [web.Server] 的工具
//
// 防止交互引用，独立为一个包。
package servertest

import (
	"crypto/tls"
	_ "embed"
	"net"
	"net/http"
	"strconv"

	"github.com/issue9/assert/v6"
)

//go:embed cert.pem
var cert []byte

//go:embed key.pem
var key []byte

type Server struct {
	URL    string
	A      *assert.Assertion
	Client *http.Client
	Port   string

	// tls 证书
	//
	// 只有在 [NewServer] 的 sec 为 true 时才会有此值
	Certificates []tls.Certificate
}

// NewServer 创建用于测试的服务数据
func NewServer(a *assert.Assertion, sec bool) *Server {
	// NOTE: 以下方式获取临时端口依然可能会出现竞态，但是概率小。

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		a.TB().Fatalf("为测试服务器分配临时端口失败：%s", err)
		return nil
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		a.TB().Fatalf("释放测试服务器临时端口失败：%s", err)
		return nil
	}

	addr := ":" + strconv.Itoa(port)

	url := "http://localhost" + addr
	if sec {
		url = "https://localhost" + addr
	}

	client := &http.Client{}
	var certs []tls.Certificate

	if sec {
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		cert, err := tls.X509KeyPair(cert, key)
		a.NotError(err).NotNil(cert)
		certs = []tls.Certificate{cert}
	}

	return &Server{
		URL:          url,
		A:            a,
		Client:       client,
		Port:         addr,
		Certificates: certs,
	}
}

// Run 异步运行 s.Serve
//
// 返回参数用于退出，在不调用返回参数的情况下，即使 s 已经中止，也不会退出。
func Run(a *assert.Assertion, s interface{ Serve() error }) func() {
	ok := make(chan struct{}, 1)
	exit := make(chan struct{}, 1)

	go func() {
		a.TB().Helper()

		defer func() { exit <- struct{}{} }()

		ok <- struct{}{} // 最起码等待协程启动
		a.ErrorIs(s.Serve(), http.ErrServerClosed)
	}()

	<-ok
	return func() { <-exit }
}
