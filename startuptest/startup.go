// SPDX-FileCopyrightText: 2026 caixw
//
// SPDX-License-Identifier: MIT

// Package startuptest 提供了一系列用于测试的 [web.Startup] 实现
package startuptest

import (
	"net/http"

	"github.com/issue9/assert/v6"
	"github.com/issue9/assert/v6/rest"

	"github.com/issue9/web"
)

// Startup 用于测试的启动器接口
type Startup interface {
	web.Startup

	// URL 服务器的访问地址
	URL() string

	// Assertion 断言对象
	Assertion() *assert.Assertion

	// Client 用于访问服务地址的客户端对象
	Client() *http.Client

	// NewRequest 创建访问请求对象
	NewRequest(method, path string) *rest.Request

	// Get 创建 Get 请求
	Get(path string) *rest.Request

	Put(path string) *rest.Request

	Post(path string) *rest.Request

	Patch(path string) *rest.Request

	Delete(path string) *rest.Request
}
