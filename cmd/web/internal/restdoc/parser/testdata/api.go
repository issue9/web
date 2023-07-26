// SPDX-License-Identifier: MIT

package testdata

// api 函数说明
//
// # api POST /login 登录
// @tag users
// @req req 登录的账号信息
// @header h1
// @cookie c1 desc
// @query name query
// @resp 201 resp
// @resp-header 201 h1 h1 desc
// @resp-header 201 h2 h2 desc
// @resp-ref 400 400-resp
// @resp-ref 404 404-resp
// @resp 200 resp resp desc
//
// 如果有其它需要详细说的，在文档最后写入，
// 会被以 md 的格式传递给 api.Description
func login() {}

type req struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type resp struct {
	UID   int    `json:"uid"`
	Token string `json:"token"`
}

type query struct {
	Type string `query:"type"`
}
