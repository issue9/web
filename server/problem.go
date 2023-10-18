// SPDX-License-Identifier: MIT

package server

import (
	"github.com/issue9/localeutil"

	"github.com/issue9/web"
)

// AddProblem 添加新的错误代码
func (srv *httpServer) AddProblem(id string, status int, title, detail web.LocaleStringer) {
	srv.problems.Add(id, status, title, detail)
}

// VisitProblems 遍历错误代码
//
// visit 签名：
//
//	func(prefix, id string, status int, title, detail LocaleStringer)
//
// prefix 用户设置的前缀，可能为空值；
// id 为错误代码，不包含前缀部分；
// status 该错误代码反馈给用户的 HTTP 状态码；
// title 错误代码的简要描述；
// detail 错误代码的明细；
func (srv *httpServer) VisitProblems(visit func(prefix, id string, status int, title, detail web.LocaleStringer)) {
	srv.problems.Visit(visit)
}

func (srv *httpServer) InitProblem(pp *web.RFC7807, id string, p *localeutil.Printer) {
	sp := srv.problems.Problem(id)
	pp.Init(sp.Type, sp.Title.LocaleString(p), sp.Detail.LocaleString(p), sp.Status)
}
