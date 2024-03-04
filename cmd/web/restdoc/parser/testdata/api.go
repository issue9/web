// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package testdata

// api 函数说明
//
// # api post /login 登录
// @tag users
// @req * req 登录的账号信息
// @header h1
// @cookie c1 desc
// @query query
// @resp 201 * resp
// @resp-header 201 h2011 h1 desc
// @resp-header 201 h2012 h2 desc
// @resp 200 * resp desc
// @security oauth-code
//
// ## callback onData POST {$request.query.url} 回调1
// @req * req 登录的账号信息
// @resp 201 * resp
//
// 如果有其它需要详细说的，在文档最后写入，
// 会被以 md 的格式传递给 api.Description
func login() {}

// 登录信息
//
// 用户登录需要提交的信息。
type req struct {
	Username string `json:"username"` // 账号
	Password string `json:"password"` // 密码
}

type resp struct {
	UID   int    `json:"uid"`
	Token string `json:"token"`
}

type query struct {
	Sex  Sex    `query:"sex"`
	Type string `query:"type"`
	X    []int
}

// @type string
// @enum male female
type Sex int
