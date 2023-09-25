// SPDX-License-Identifier: MIT

// Package micro 提供微服务相关功能
package micro

import (
	"io"
	"net/http"

	"github.com/issue9/web"
)

// 不采用 http.DefaultClient，防止无意中被修改。
var defaultClient = &http.Client{}

// TODO: 将 web.Server 独立为接口，同时供 web.Server,micro.Node,micro.Gateway 使用？

// 以 API 的形式调用微服务
func requestAPI[T any](method, path string, input io.Reader, resp *T, problem *web.RFC7807) error {
	req, err := http.NewRequest(method, path, input)
	if err != nil {
		return err
	}

	r, err := defaultClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	// TODO parse content-encoding,content-type
	// web.Server.Mimetypes().contentType(...)
	var unmarshal web.UnmarshalFunc

	if r.StatusCode >= 400 {
		return unmarshal(data, problem)
	}
	return unmarshal(data, resp)
}
