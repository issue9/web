// SPDX-License-Identifier: MIT

// Package testdata 测试数据
//
// 这是测试数据的说明
//
// # restdoc RESTDoc 标题
//
// @tag admin admin API
// @tag users users API
// @server https://api.example.com/v1 v1 api
// @server https://api.example.com/v2 v2 api
// @license mit https://license.example.com/mit
// @term https://term.example.com
// @version 1.0.0
// @media application/json application/xml
// @resp 400-resp resp400 400 错误
// @resp 404-resp resp404 not found
// @resp-media 400-resp application/problem+json application/problem+xml
// @resp-media 404-resp application/problem+json application/problem+xml
//
// # 其它文档说明
//
// 这也将被传递维给 info.Description
package testdata

type resp400 struct {
	Status int `json:"status" xml:"status,attr"`
}

type resp404 struct {
	Status int `json:"status" xml:"status,attr"`
}
