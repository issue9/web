// SPDX-License-Identifier: MIT

// Package testdata 测试数据
//
// 这是测试数据的说明
//
// # restdoc RESTDoc 标题
//
// @tag admin admin API
// @tag users users API
// @server * https://api.example.com/v1 v1 api
// @server * https://api.example.com/v2 v2 api
// @license mit https://license.example.com/mit
// @term https://term.example.com
// @version 1.0.0
// @media application/json application/xml
// @header h1 自定义报头
// @resp 400 application/problem+json,application/problem+xml resp400 400 错误
// @resp 404 application/problem+json,application/problem+xml resp404 not found
//
// @scy-code oauth-code https://example.com/auth https://example.com/token https://example.com/refresh read:info,write:info
//
// # 其它文档说明
//
// 这也将被传递维给 info.Description
package testdata

type resp400 struct {
	XMLName struct{} `json:"-" xml:"resp-400"`
	Status  int      `json:"status" xml:"status,attr"`
}

type resp404 struct {
	Status int `json:"status" xml:"status,attr"`
}
